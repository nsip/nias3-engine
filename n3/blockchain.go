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
	author  []byte
	tip     []byte
	db      *bolt.DB
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	context     string
	author      []byte
	currentHash []byte
	db          *bolt.DB
}

//
// AddBlock saves provided data as a block in the blockchain
// returns the accpted block
//
func (bc *Blockchain) AddBlock(data *SPOTuple) *Block {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {

		root := tx.Bucket([]byte(contextsBucket))
		cntx := root.Bucket([]byte(data.Context))
		usr := cntx.Bucket([]byte(cs.PublicID()))
		b := usr.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Fatal("add-block error: ", err)
	}

	newBlock, err := NewBlock(data, lastHash)
	if err != nil {
		log.Println("unable to add block: ", err)
		return nil
	}

	err = bc.db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(contextsBucket))
		cntx := root.Bucket([]byte(data.Context))
		usr := cntx.Bucket([]byte(cs.PublicID()))
		b := usr.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})

	return newBlock
}

// Iterator ...
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{currentHash: bc.tip,
		db:      bc.db,
		author:  bc.author,
		context: bc.context}

	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(contextsBucket))
		cntx := root.Bucket([]byte(i.context))
		usr := cntx.Bucket(i.author)
		b := usr.Bucket([]byte(blocksBucket))

		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}

//
// NewBlockchain creates a new Blockchain with genesis Block
// for the current owner with the specified context
//
func NewBlockchain(contextName string) *Blockchain {
	var tip []byte
	db := boltDB
	author := []byte(cs.PublicID())

	err := db.Update(func(tx *bolt.Tx) error {

		root, err := tx.CreateBucketIfNotExists([]byte(contextsBucket))
		cntx, err := root.CreateBucketIfNotExists([]byte(contextName))
		usr, err := cntx.CreateBucketIfNotExists(author)
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

			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
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
func GetBlockchain(contextName string, author []byte) *Blockchain {

	var tip []byte
	db := boltDB

	err := db.Update(func(tx *bolt.Tx) error {

		root, err := tx.CreateBucketIfNotExists([]byte(contextsBucket))
		cntx, err := root.CreateBucketIfNotExists([]byte(contextName))
		usr, err := cntx.CreateBucketIfNotExists(author)
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
