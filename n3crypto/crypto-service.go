// crypto-service.go

package n3crypto

import (
	"log"
	"time"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/nuid"
	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
)

// Ensure CryptoService implements n3.CryptoService interface
var _ n3.CryptoService = &CryptoService{}

type CryptoService struct {
	cryptoClient *cryptoClient
}

func NewCryptoService() (*CryptoService, error) {
	cs := &CryptoService{}
	cc, err := NewCryptoClient()
	if err != nil {
		return nil, err
	}
	cs.cryptoClient = cc
	// log.Println("crypto service created...")
	return cs, nil

}

//
// returns the public id of the underlying n3 instance
// p2p nodes have their own identity
//
func (cs *CryptoService) PublicID() (string, error) {

	idFromKey, err := peer.IDFromPublicKey(cs.cryptoClient.pubKey)
	if err != nil {
		log.Println(err, "Failed to extract peer id from public key")
		return "", err
	}
	return idFromKey.Pretty(), nil
}

//
// validates messages against included signature
//
func (cs *CryptoService) AuthenticateMessage(msg *messages.SignedN3Message) (bool, error) {

	// marshall data to protobufs3 binary format
	bin, err := proto.Marshal(msg.N3Message)
	if err != nil {
		return false, err
	}

	// restore peer id binary format from base58 encoded node id data
	nodeId := msg.N3Message.MessageData.NodeId
	peerId, err := peer.IDB58Decode(nodeId)
	if err != nil {
		return false, err
	}
	// get public key of sending node
	nodePubKey := msg.N3Message.MessageData.NodePubKey

	// verify the data was authored by the signing peer identified by the public key
	// and signature included in the message
	return cs.verifyData(bin, []byte(msg.Signature), peerId, nodePubKey), nil

}

// Verify incoming p2p message data integrity
// data: data to verify
// signature: author signature provided in the message payload
// peerId: author peer id from the message payload
// pubKeyData: author public key from the message payload
func (cs *CryptoService) verifyData(data []byte, signature []byte, peerId peer.ID, pubKeyData []byte) bool {
	key, err := crypto.UnmarshalPublicKey(pubKeyData)
	if err != nil {
		log.Println(err, "Failed to extract key from message key data")
		return false
	}

	// extract node id from the provided public key
	idFromKey, err := peer.IDFromPublicKey(key)

	if err != nil {
		log.Println(err, "Failed to extract peer id from public key")
		return false
	}

	// verify that message author node id matches the provided node public key
	if idFromKey != peerId {
		log.Println(err, "Node id and provided public key mismatch")
		return false
	}

	res, err := key.Verify(data, signature)
	if err != nil {
		log.Println(err, "Error authenticating data")
		return false
	}

	return res
}

//
// generates a message meta-data block - does not depend on message
// content
//
func (cs *CryptoService) NewMessageData(gossip bool) (*messages.MessageData, error) {

	idFromKey, err := peer.IDFromPublicKey(cs.cryptoClient.pubKey)
	if err != nil {
		log.Println(err, "Failed to extract peer id from public key")
		return nil, err
	}

	nodeId := peer.IDB58Encode(idFromKey)
	nodePubKey, err := cs.cryptoClient.pubKey.Bytes()
	if err != nil {
		return nil, err
	}

	return &messages.MessageData{ClientVersion: n3.N3ClientVersion,
		NodeId:     nodeId,
		NodePubKey: nodePubKey,
		Timestamp:  time.Now().Unix(),
		Id:         nuid.Next(),
		Gossip:     gossip}, nil

}

//
// Creates a digital signature for the bytes of the provided message
// with the local private key
// return a signature based on the message content
//
func (cs *CryptoService) SignMessage(msg *messages.N3Message) (*messages.SignedN3Message, error) {

	msgData, err := msg.ToBytes()
	if err != nil {
		return nil, err
	}
	// generate the signature
	signature, err := cs.signMessage(msgData)
	// add signature to outer data wrapper
	signed_msg := &messages.SignedN3Message{N3Message: msg, Signature: string(signature)}

	return signed_msg, nil

}

//
// internal method that actually does the signing
//
func (cs *CryptoService) signMessage(data []byte) ([]byte, error) {
	key := privKey
	res, err := key.Sign(data)
	return res, err

}
