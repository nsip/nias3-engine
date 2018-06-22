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
	"strconv"
	"strings"
	"text/scanner"

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

// We quote the components of both the tuple and the key, to prevent ambiguity
func (t *SPOTuple) Bytes() []byte {
	return []byte(fmt.Sprintf("%s:%s:%s:%s:%d", strconv.Quote(t.Context), strconv.Quote(t.Subject), strconv.Quote(t.Predicate), strconv.Quote(t.Object), t.Version))
}

//
// helper method provides lookup key for use with
// count-min-sketch for tuple versioning

// All keys quote their values, to avoid ambiguity with space as a key part delimiter, and terminate in space, so that a third key component prefix can be differentiated from a complete third key component value

// This key is SPO
func (t *SPOTuple) CmsKey() string {
	return fmt.Sprintf("c:%s s:%s p:%s o:%s ", strconv.Quote(t.Context), strconv.Quote(t.Subject), strconv.Quote(t.Predicate), strconv.Quote(t.Object))
}

func (t *SPOTuple) CmsKeySP() string {
	return fmt.Sprintf("c:%s s:%s p:%s ", strconv.Quote(t.Context), strconv.Quote(t.Subject), strconv.Quote(t.Predicate))
}

func (t *SPOTuple) CmsKeySPO() string {
	return fmt.Sprintf("c:%s s:%s p:%s o:%s ", strconv.Quote(t.Context), strconv.Quote(t.Subject), strconv.Quote(t.Predicate), strconv.Quote(t.Object))
}

func (t *SPOTuple) CmsKeySOP() string {
	return fmt.Sprintf("c:%s s:%s o:%s p:%s ", strconv.Quote(t.Context), strconv.Quote(t.Subject), strconv.Quote(t.Object), strconv.Quote(t.Predicate))
}

func (t *SPOTuple) CmsKeyPSO() string {
	return fmt.Sprintf("c:%s p:%s s:%s o:%s ", strconv.Quote(t.Context), strconv.Quote(t.Predicate), strconv.Quote(t.Subject), strconv.Quote(t.Object))
}

func (t *SPOTuple) CmsKeyPOS() string {
	return fmt.Sprintf("c:%s p:%s o:%s s:%s ", strconv.Quote(t.Context), strconv.Quote(t.Predicate), strconv.Quote(t.Object), strconv.Quote(t.Subject))
}

func (t *SPOTuple) CmsKeyOPS() string {
	return fmt.Sprintf("c:%s s:%s o:%s p:%s ", strconv.Quote(t.Context), strconv.Quote(t.Object), strconv.Quote(t.Predicate), strconv.Quote(t.Subject))
}

func (t *SPOTuple) CmsKeyOSP() string {
	return fmt.Sprintf("c:%s s:%s o:%s p:%s ", strconv.Quote(t.Context), strconv.Quote(t.Object), strconv.Quote(t.Subject), strconv.Quote(t.Predicate))
}

//
// helper for proto encoding
//
func (t *SPOTuple) Serialize() []byte {

	out, err := proto.Marshal(t)
	if err != nil {
		log.Println("tuple-serialize: protobuf encoding error: ", err)
	}
	return out

}

//
// helper for proto decoding
//
func DeserializeTuple(d []byte) *SPOTuple {

	tuple := &SPOTuple{}

	err := proto.Unmarshal(d, tuple)
	if err != nil {
		log.Println("tuple-deserialize: protobuf decoding error: ", err)
	}

	return tuple

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

// parse a tuple key into a triple
func ParseTripleKey(t []byte) SPOTuple {
	var s scanner.Scanner
	s.Init(strings.NewReader(string(t)))
	o := SPOTuple{}
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if s.TokenText() == "c" {
			tok = s.Scan() // colon
			tok = s.Scan()
			o.Context, _ = strconv.Unquote(s.TokenText())
		}
		if s.TokenText() == "s" {
			tok = s.Scan() // colon
			tok = s.Scan()
			o.Subject, _ = strconv.Unquote(s.TokenText())
		}
		if s.TokenText() == "p" {
			tok = s.Scan() // colon
			tok = s.Scan()
			o.Predicate, _ = strconv.Unquote(s.TokenText())
		}
		if s.TokenText() == "o" {
			tok = s.Scan() // colon
			tok = s.Scan()
			o.Object, _ = strconv.Unquote(s.TokenText())
		}
	}
	return o
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
