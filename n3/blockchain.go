// blockchain.go

package n3

import (
	"fmt"
	"log"

	"github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const contextsBucket = "contexts"

var boltDB *bolt.DB

func init() {
	if boltDB == nil {
		var dbErr error
		dbPath := "./n3.db"
		boltDB, dbErr = bolt.Open("n3.db", 0600, nil)
		if dbErr != nil {
			log.Fatal(errors.Wrap(dbErr, "cannot open n3 datastore "+dbPath))
		}
	}
}

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
	context string
	author  string
	tip     []byte
	db      *bolt.DB
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

		// fmt.Println("lookup a-n-b:")
		// DeserializeBlock(b.Get([]byte(lastHash))).Print()

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "AddNewBlock error: ")
	}

	// create the new block as next in chain
	// log.Println("\t\tlast hash: ", lastHash)
	newBlock, err := NewBlock(data, lastHash)
	if err != nil {
		return nil, errors.Wrap(err, "AddNewBlock error: ")
	}

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
	if err != nil {
		return nil, errors.Wrap(err, "AddBlock error: ")
	}

	return newBlock, nil

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

	// log.Println("previous:")
	// prevBlock.Print()
	// log.Println("new:")
	// b.Print()

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
// NewBlockchain creates a new Blockchain with genesis Block
// for the current owner with the specified context
//
func NewBlockchain(contextName string) *Blockchain {
	var tip []byte
	db := boltDB
	author := cs.PublicID()

	err := db.Update(func(tx *bolt.Tx) error {

		root, err := tx.CreateBucketIfNotExists([]byte(contextsBucket))
		cntx, err := root.CreateBucketIfNotExists([]byte(contextName))
		usr, err := cntx.CreateBucketIfNotExists([]byte(author))
		if err != nil {
			return err
		}
		b := usr.Bucket([]byte(blocksBucket))

		if b == nil {
			fmt.Println("No existing blockchain found. Creating a new one...")
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

	bc := Blockchain{context: contextName,
		author: author,
		tip:    tip,
		db:     db}

	return &bc
}

//
// returns the blockchain for the given context/user if one exists,
// otherwise retuns an empty blockchain container
//
func GetBlockchain(contextName string, author string) *Blockchain {

	var tip []byte
	db := boltDB

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

	bc := &Blockchain{context: contextName,
		author: author,
		tip:    tip,
		db:     db}

	return bc

}
