// blockchain.go

package n3

import (
	"log"

	"github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

const blocksBucket = "blocks"
const contextsBucket = "contexts"

var boltDB *bolt.DB

var localCMSStore map[string]*N3CMS

func init() {
	if boltDB == nil {
		var dbErr error
		boltDB, dbErr = bolt.Open("n3.db", 0600, &bolt.Options{NoFreelistSync: true})
		if dbErr != nil {
			log.Fatal(errors.Wrap(dbErr, "cannot open n3 blockchain datastore."))
		}
	}
	localCMSStore = make(map[string]*N3CMS)
}

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
	context string
	author  string
	tip     []byte
	db      *bolt.DB
	cms     *N3CMS // manages tuple versioning for this b/c
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	context     string
	author      string
	currentHash []byte
	db          *bolt.DB
}

//
// AddNewBlock creates a block in the blockchain
// from the tuple provided
// returns the accpted block, or an error if the
// block is not valid
//
func (bc *Blockchain) AddNewBlock(data *SPOTuple) (*Block, error) {

	var lastHash string

	// find current tip of the chain
	err := bc.db.View(func(tx *bolt.Tx) error {

		root := tx.Bucket([]byte(contextsBucket))
		cntx := root.Bucket([]byte(data.Context))
		usr := cntx.Bucket([]byte(cs.PublicID()))
		b := usr.Bucket([]byte(blocksBucket))
		lastHash = string(b.Get([]byte("l")))

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "AddNewBlock error: ")
	}

	// assign data tuple version within this b/c context
	cmsKey := data.CmsKeySP()
	knownVer := bc.cms.Estimate(cmsKey)
	if knownVer > 0 {
		nextVer := knownVer + 1
		data.Version = nextVer
		bc.cms.Update(cmsKey, 1)
	} else {
		data.Version = 1
		bc.cms.Update(cmsKey, 1)
	}

	// create the new block as next in chain
	newBlock, err := NewBlock(data, lastHash)
	if err != nil {
		return nil, errors.Wrap(err, "AddNewBlock error: ")
	}
	// log.Printf("\n\ttuple: %+v\n\tversion:%d\n", newBlock.Data, newBlock.Data.Version)

	/*
		// add the block to the chain
		addedBlock, err := bc.AddBlock(newBlock)
		if err != nil {
			return nil, errors.Wrap(err, "AddNewBlock error: ")
		}

		// update the tip
		err = bc.db.Update(func(tx *bolt.Tx) error {
			root := tx.Bucket([]byte(contextsBucket))
			cntx := root.Bucket([]byte(data.Context))
			usr := cntx.Bucket([]byte(cs.PublicID()))
			b := usr.Bucket([]byte(blocksBucket))
			err = b.Put([]byte("l"), []byte(addedBlock.Hash))
			if err != nil {
				log.Println("db l update error:", err)
				return err
			}

			bc.tip = []byte(addedBlock.Hash)

			return nil
		})
	*/
	// FOLD INTO ONE TRANSACTION
	//var prevBlock *Block

	err = bc.db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(contextsBucket))
		cntx := root.Bucket([]byte(data.Context))
		usr := cntx.Bucket([]byte(cs.PublicID()))
		bkt := usr.Bucket([]byte(blocksBucket))
		/* redundant: we have set newBlock.PrevBlockHash to be prevBlock.Hash already! */
		/*
			prevBlock = DeserializeBlock(bkt.Get([]byte(newBlock.PrevBlockHash)))
			if prevBlock.Hash != newBlock.PrevBlockHash {
				return errors.New("attempted to add invalid block: block has no chain.")
			}
		*/
		err := bkt.Put([]byte(newBlock.Hash), newBlock.Serialize())
		if err != nil {
			log.Println("db l update error:", err)
			return err
		}
		err = bkt.Put([]byte("l"), []byte(newBlock.Hash))
		if err != nil {
			log.Println("db l update error:", err)
			return err
		}

		bc.tip = []byte(newBlock.Hash)

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "AddBlock error: ")
	}

	return newBlock, nil

}

// AddNewBlocks creates multiple blocks in the blockchain
// from the tuples provided
// returns the accpted block, or an error if the
// block is not valid
// Assumes all data belongs to the same context
//
func (bc *Blockchain) AddNewBlocks(datablocks []*SPOTuple) ([]*Block, error) {

	var lastHash string
	ret := make([]*Block, 0)

	err := bc.db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(contextsBucket))
		cntx := root.Bucket([]byte(datablocks[0].Context))
		usr := cntx.Bucket([]byte(cs.PublicID()))
		bkt := usr.Bucket([]byte(blocksBucket))
		// find current tip of the chain
		lastHash = string(bkt.Get([]byte("l")))

		for _, data := range datablocks {

			// assign data tuple version within this b/c context
			cmsKey := data.CmsKeySP()
			knownVer := bc.cms.Estimate(cmsKey)
			if knownVer > 0 {
				nextVer := knownVer + 1
				data.Version = nextVer
				bc.cms.Update(cmsKey, 1)
			} else {
				data.Version = 1
				bc.cms.Update(cmsKey, 1)
			}

			// create the new block as next in chain
			newBlock, err1 := NewBlock(data, lastHash)
			if err1 != nil {
				return errors.Wrap(err1, "AddNewBlock error: ")
			}
			// log.Printf("\n\ttuple: %+v\n\tversion:%d\n", newBlock.Data, newBlock.Data.Version)

			err1 = bkt.Put([]byte(newBlock.Hash), newBlock.Serialize())
			if err1 != nil {
				log.Println("db l update error:", err1)
				return err1
			}
			lastHash = newBlock.Hash
			ret = append(ret, newBlock)
		}
		err1 := bkt.Put([]byte("l"), []byte(lastHash))
		if err1 != nil {
			log.Println("db l update error:", err1)
			return err1
		}
		bc.tip = []byte(lastHash)

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "AddBlock error: ")
	}

	return ret, nil
}

//
// Does the work of checking the validity of the block
// and adding it to the blockchain
// can be used on its own to validate blocks from other users
//
func (bc *Blockchain) AddBlock(b *Block) (*Block, error) {

	var prevBlock *Block

	// validate block by getting predecessor
	err := bc.db.View(func(tx *bolt.Tx) error {

		root := tx.Bucket([]byte(contextsBucket))
		cntx := root.Bucket([]byte(b.Data.Context))
		usr := cntx.Bucket([]byte(b.Author))
		bkt := usr.Bucket([]byte(blocksBucket))
		prevBlock = DeserializeBlock(bkt.Get([]byte(b.PrevBlockHash)))

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "AddBlock tx error: ")
	}

	if prevBlock.Hash != b.PrevBlockHash {
		return nil, errors.New("attempted to add invalid block: block has no chain.")
	}

	// assuming valid add to the blockchain
	err = bc.db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(contextsBucket))
		cntx := root.Bucket([]byte(b.Data.Context))
		usr := cntx.Bucket([]byte(b.Author))
		bkt := usr.Bucket([]byte(blocksBucket))
		err := bkt.Put([]byte(b.Hash), b.Serialize())
		if err != nil {
			return err
		}

		bc.tip = []byte(b.Hash)

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "AddBlock tx error: ")
	}

	return b, nil

}

// Iterator ...
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{
		currentHash: bc.tip,
		db:          bc.db,
		author:      bc.author,
		context:     bc.context,
	}

	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(contextsBucket))
		cntx := root.Bucket([]byte(i.context))
		usr := cntx.Bucket([]byte(i.author))
		b := usr.Bucket([]byte(blocksBucket))

		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Println("iterator error:", err)
	}

	i.currentHash = []byte(block.PrevBlockHash)

	return block
}

//
// cleanly tidy up any stateful resources
//
func (bc *Blockchain) Close() {
	bc.db.Close()
	bc.cms.Close()
}

//
// NewBlockchain creates a new Blockchain with genesis Block
// for the current owner with the specified context
//
func NewBlockchain(contextName string) *Blockchain {
	var tip []byte
	db := boltDB
	author := cs.PublicID()

	fileName := "./" + contextName + ".cms"
	cms, cmsErr := NewN3CMS(fileName)
	if cmsErr != nil {
		log.Fatal("unable to create b/c version cms: ", cmsErr)
	}
	localCMSStore[contextName] = cms

	err := db.Update(func(tx *bolt.Tx) error {

		root, err := tx.CreateBucketIfNotExists([]byte(contextsBucket))
		cntx, err := root.CreateBucketIfNotExists([]byte(contextName))
		usr, err := cntx.CreateBucketIfNotExists([]byte(author))
		if err != nil {
			return err
		}
		b := usr.Bucket([]byte(blocksBucket))

		if b == nil {
			log.Println("No existing blockchain found. Creating a new one...")
			genesis, err := NewGenesisBlock(contextName)
			if err != nil {
				return err
			}

			b, err := usr.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}

			err = b.Put([]byte(genesis.Hash), genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = b.Put([]byte("l"), []byte(genesis.Hash))
			if err != nil {
				log.Panic(err)
			}
			tip = []byte(genesis.Hash)
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{
		context: contextName,
		author:  author,
		tip:     tip,
		db:      db,
		cms:     cms,
	}

	return &bc
}

//
// returns the blockchain for the given context/user if one exists,
// otherwise returns an empty blockchain container. If the blockchain
// is local to the node (being created to process local requests:
// author = cs.PublicID() ),
// will create new local blockchain if one does not already exist.
//
func GetBlockchain(contextName string, author string) *Blockchain {

	var tip []byte
	db := boltDB

	cms, ok := localCMSStore[contextName]
	if author != cs.PublicID() {
		cms = nil
	} else if !ok {
		return NewBlockchain(contextName)
	}

	err := db.Update(func(tx *bolt.Tx) error {

		root, err := tx.CreateBucketIfNotExists([]byte(contextsBucket))
		cntx, err := root.CreateBucketIfNotExists([]byte(contextName))
		usr, err := cntx.CreateBucketIfNotExists([]byte(author))
		b, err := usr.CreateBucketIfNotExists([]byte(blocksBucket))
		if err != nil {
			return err
		}

		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := &Blockchain{
		context: contextName,
		author:  author,
		tip:     tip,
		db:      db,
		cms:     cms,
	}

	return bc

}
