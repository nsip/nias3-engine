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
	"strconv"
	"time"
)

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
func NewBlock(data *SPOTuple, prevBlockHash []byte) *Block {
	block := &Block{Timestamp: time.Now().Unix(),
		Data:          data,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Sig:           []byte{},
		Author:        []byte{}, // set this from cs
	}

	// assign tuple version
	// assignTupleVersion()

	// assign author - id from this machine
	block.Author = []byte("itsme")

	// assign new hash
	block.setHash()

	// now sign the completed block
	block.sign()

	return block
}

func (b *Block) setHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data.Bytes(), timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

func (b *Block) sign() {
	b.Sig = []byte("i have been signed")
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(contextName string) *Block {
	t := &SPOTuple{Subject: "Genesis",
		Predicate: "Genesis",
		Object:    "Genesis",
		Context:   contextName}
	return NewBlock(t, []byte{})
}

// Serialize serializes the block
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
