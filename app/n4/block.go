// block.go

package n4

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64
	Data          SPOTuple
	PrevBlockHash []byte
	Hash          []byte
	Sig           []byte
	Author        []byte
}

type SPOTuple struct {
	Context   string
	Subject   string
	Predicate string
	Object    string
	Version   int
}

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
func NewBlock(data SPOTuple, prevBlockHash []byte) *Block {
	block := &Block{Timestamp: time.Now().Unix(),
		Data:          data,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Sig:           []byte{},
		Author:        []byte{}, // set this from cs
	}

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
	t := SPOTuple{Subject: "Genesis",
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
