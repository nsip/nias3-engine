// config.go

package n3stan

import (
	"log"

	"github.com/nsip/nias3-engine/n3crypto"
	"github.com/nsip/nias3-engine/utils"
)

var cs, _ = n3crypto.NewCryptoService()
var portString string

type SubscriptionConfig struct {
	Topic       string
	DurableName string
}

type StanConfig struct {
	Url       string
	NatsPort  string
	ClientId  string
	ClusterId string
}

func DefaultStanConfig(clientID string) StanConfig {

	pID, err := cs.PublicID()
	if err != nil {
		log.Println("error fetching public id from crypto service, temp cid assigned.")
		pID = "TEMP"
	}

	if portString == "" {
		portString = utils.GetAvailPortString()
	}

	return StanConfig{
		Url:       "nats://localhost:" + portString,
		NatsPort:  portString,
		ClientId:  clientID,
		ClusterId: "n3" + pID,
	}
}
