// crypto-service.go

package n3crypto

import (
	"encoding/hex"
	"log"

	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
)

type CryptoService struct {
	cryptoClient *cryptoClient
}

func NewCryptoService() *CryptoService {
	cs := &CryptoService{}
	cc, err := NewCryptoClient()
	if err != nil {
		log.Fatal("cannot create crypto service: ", err)
	}
	cs.cryptoClient = cc

	return cs

}

//
// returns the public id of the underlying n3 instance
// p2p nodes have their own identity
//
func (cs *CryptoService) PublicID() string {

	idFromKey, err := peer.IDFromPublicKey(cs.cryptoClient.pubKey)
	if err != nil {
		log.Fatalln("Failed to extract peer id from public key: ", err)
	}
	return idFromKey.Pretty()
}

//
// sign a block
//
func (cs *CryptoService) SignBlock(blockData []byte) ([]byte, error) {
	key := privKey
	res, err := key.Sign(blockData)
	return res, err
}

//
// verify a signed block was created as author intended
//
func (cs *CryptoService) VerifyBlock(blockData []byte, blockAuthor string, blockSig string) (bool, error) {

	// decode block sig from hex
	sigBytes, err := hex.DecodeString(blockSig)
	if err != nil {
		return false, errors.Wrap(err, "cannot hex decode block signature")
	}

	// // restore peer id binary format from base58 encoded node id data
	authorId, err := peer.IDB58Decode(blockAuthor)
	if err != nil {
		return false, err
	}

	key, err := authorId.ExtractPublicKey()
	if err != nil {
		return false, errors.Wrap(err, "Failed to extract key from block author")
	}

	res, err := key.Verify(blockData, sigBytes)
	if err != nil {
		return false, errors.Wrap(err, "Error authenticating data")
	}

	return res, nil

}
