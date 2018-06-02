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
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
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
//  Sender		  []byte
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
// helper method provides lookup key for use with
// count-min-sketch for tuple versioning
//
func (t *SPOTuple) CmsKey() string {
	return fmt.Sprintf("%s:%s:%s:%s", t.Context, t.Subject, t.Predicate, t.Object)
}

//
// check signature against author & content
//
func (b *Block) Verify() bool {

	isVerified, err := cs.VerifyBlock(b.signablePayload(), b.Author, b.Sig)
	if err != nil {
		log.Println("block verification failed: ", err)
	}

	return isVerified
}

// NewBlock creates and returns Block
func NewBlock(data *SPOTuple, prevBlockHash string) (*Block, error) {

	block := &Block{
		BlockId:       nuid.Next(),
		Data:          data,
		PrevBlockHash: prevBlockHash,
		Hash:          "",
		Sig:           "",
		Author:        cs.PublicID(),
		Sender:        cs.PublicID(),
	}

	// assign tuple version
	// assignTupleVersion()

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
	headers := bytes.Join([][]byte{
		[]byte(b.PrevBlockHash),
		b.Data.Bytes(),
		[]byte(b.BlockId),
	}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = fmt.Sprintf("%x", hash[:])
}

//
// add crypto signature to the payload
//
func (b *Block) sign() error {
	var sigErr error
	signBytes := b.signablePayload()
	sig, sigErr := cs.SignBlock(signBytes)
	if sigErr != nil {
		return sigErr
	}

	b.Sig = fmt.Sprintf("%x", sig)

	return nil
}

//
// returns the parts of the block content used for
// signature validation
//
func (b *Block) signablePayload() []byte {

	elements := [][]byte{
		[]byte(b.BlockId),
		[]byte(b.Author),
		[]byte(b.Hash),
		[]byte(b.PrevBlockHash),
		b.Data.Bytes(),
	}
	return bytes.Join(elements, []byte{})

}

func (b *Block) CmsKey() string {
	return fmt.Sprintf("%s:%s", b.BlockId, b.Receiver)
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(contextName string) (*Block, error) {
	t := &SPOTuple{Subject: "Genesis",
		Predicate: "Genesis",
		Object:    "Genesis",
		Context:   contextName,
		Version:   0,
	}
	blk, err := NewBlock(t, "")
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

	out, err := proto.Marshal(b)
	if err != nil {
		log.Println("block-serialize: protobuf encoding error: ", err)
	}
	return out

}

//
// pretty-print a block
//
func (b *Block) Print() {
	log.Printf(`

		Id:
		%s 
		Data:
		%+v 
		PrevBlockHash:
		%s
		Hash:
		%s
		Sig:
		%s
		Author:        
		%s
		Sender:        
		%s 
		Receiver:
		%s

`, b.BlockId, b.Data, b.PrevBlockHash, b.Hash, b.Sig, b.Author, b.Sender, b.Receiver)
}

//
// DeserializeBlock deserializes a block from the datastore
// TODO: change to protobuf encoding
//
func DeserializeBlock(d []byte) *Block {

	block := &Block{}

	err := proto.Unmarshal(d, block)
	if err != nil {
		log.Println("block-deserialize: protobuf decoding error: ", err)
	}

	return block

}
