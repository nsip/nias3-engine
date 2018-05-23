// crypto-client.go

package n3crypto

import (
	"io/ioutil"
	"log"

	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/pkg/errors"
)

//
// crypto client maintains public/private keypair for use
// in p2p identity and message validation
//

var privKey crypto.PrivKey
var pubKey crypto.PubKey

type cryptoClient struct {
	privKey crypto.PrivKey
	pubKey  crypto.PubKey
}

//
// creates a new crypto client instance, pub/priv keys
// are created if not in existence.
//
func NewCryptoClient() (*cryptoClient, error) {

	cc := &cryptoClient{}

	if privKey == nil || pubKey == nil {
		// see if an identity was persisted
		var loadErr error
		privKey, pubKey, loadErr = loadKeys()
		// if not generate new identity for this node
		if loadErr != nil {
			var genError error
			privKey, pubKey, genError = crypto.GenerateKeyPair(crypto.Secp256k1, 256)
			if genError != nil {
				return nil, errors.Wrap(genError, "unable to generate secure keypair")
			}
		}
	}
	cc.privKey = privKey
	cc.pubKey = pubKey

	err := saveKeys(cc.privKey, cc.pubKey)
	if err != nil {
		log.Println("crypto: unable to persist keypair, identity is ephemoral")
	}

	return cc, nil

}

func loadKeys() (crypto.PrivKey, crypto.PubKey, error) {
	privKeyBytes, err := ioutil.ReadFile("prv.key")
	if err != nil {
		return nil, nil, err
	}
	pubKeyBytes, err := ioutil.ReadFile("pub.key")
	if err != nil {
		return nil, nil, err
	}
	privK, err := crypto.UnmarshalPrivateKey(privKeyBytes)
	if err != nil {
		return nil, nil, err
	}
	pubK, err := crypto.UnmarshalPublicKey(pubKeyBytes)
	if err != nil {
		return nil, nil, err
	}
	return privK, pubK, nil
}

func saveKeys(privKey crypto.PrivKey, pubKey crypto.PubKey) error {
	privKeyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return err
	}
	pubKeyBytes, err := crypto.MarshalPublicKey(pubKey)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("prv.key", privKeyBytes, 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("pub.key", pubKeyBytes, 0644)
	if err != nil {
		return err
	}
	return nil

}
