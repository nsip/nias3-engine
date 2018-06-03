// config.go

package n3

import (
	"github.com/nsip/nias3-engine/utils"
)

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

	pID := string(cs.PublicID())

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
