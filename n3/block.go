// block.go

package n3

// DO NOT remove the line below, this is used by the
// go generate tool to create the protobuf message support classes
// for all data messages.
//
//go:generate protoc --go_out=. block.proto
//

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/nats-io/nuid"
	"github.com/nsip/nias3-engine/n3crypto"
)

var cs = n3crypto.NewCryptoService()

//
// for reference only, real message definitions
// are in the protobuf files bloc.proto / block.pb.go
//
// type Block struct {
// 	Timestamp     int64
// 	Data          *SPOTuple
// 	PrevBlockHash []byte
// 	Hash          []byte
// 	Sig           []byte
// 	Author        []byte
// }

// type SPOTuple struct {
// 	Context   string
// 	Subject   string
// 	Predicate string
// 	Object    string
// 	Version   uint64
// }

func (t *SPOTuple) Bytes() []byte {
	return []byte(fmt.Sprintf("%s:%s:%s:%s:%d", t.Context, t.Subject, t.Predicate, t.Object, t.Version))
}

//
// check signature against author & content
//
func (b *Block) Verify() bool {
	return true
}

// NewBlock creates and returns Block
func NewBlock(data *SPOTuple, prevBlockHash []byte) (*Block, error) {
	block := &Block{BlockId: []byte(nuid.Next()),
		Data:          data,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Sig:           []byte{},
		Author:        []byte{}, // set this from cs
	}

	// assign tuple version
	// assignTupleVersion()

	// assign author - id from this machine
	block.Author = cs.PublicID()

	// assign new hash
	block.setHash()

	// now sign the completed block
	err := block.sign()
	if err != nil {
		log.Println("unable to sign block: ", err)
		return nil, err
	}

	return block, nil
}

func (b *Block) setHash() {
	id := b.BlockId
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data.Bytes(), id}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

func (b *Block) sign() error {
	var sigErr error
	b.Sig, sigErr = cs.SignBlock(b.Serialize())
	if sigErr != nil {
		return sigErr
	}
	return nil
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(contextName string) (*Block, error) {
	t := &SPOTuple{Subject: "Genesis",
		Predicate: "Genesis",
		Object:    "Genesis",
		Context:   contextName}
	blk, err := NewBlock(t, []byte{})
	if err != nil {
		return nil, err
	}
	return blk, nil
}

//
// Serialize serializes the block for data storage
// TODO: change to protobuf encoding
//
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

//
// DeserializeBlock deserializes a block from the datastore
// TODO: change to protobuf encoding
//
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
