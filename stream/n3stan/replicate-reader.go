// replicate-reader.go

package n3stan

import (
	"log"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
)

// Ensure IngestReader implements n3.ReaderService interface
var _ n3.ReaderService = &ReplicateReader{}

type ReplicateReader struct {
	client *readerClient
	out    <-chan *messages.SignedN3Message
}

func NewReplicateReader() n3.ReaderService {
	return &ReplicateReader{}
}

func (rr *ReplicateReader) Subscribe(topic string, clientID string) error {

	var err error

	if clientID == "" {
		return errors.New("clientID cannot be empty string")
	}

	stanConf := DefaultStanConfig(clientID)
	rr.client, err = NewReaderClient(stanConf)
	if err != nil {
		return errors.Wrap(err, "unable to create stan connection")
	}

	subConf := SubscriptionConfig{Topic: topic, DurableName: stanConf.ClientId}
	err = rr.client.subscribe(subConf)
	if err != nil {
		return errors.Wrap(err, "unable to create replicate subscription")
	}

	rr.out = make(chan *messages.SignedN3Message)
	rr.out, err = createOrderedTransportReadAdaptor(rr.client.MsgC, rr.client.ErrC)
	if err != nil {
		errors.Wrap(err, "unable to create ordered tranport adaptor")
	}

	log.Println("ReplicateReader successfully connected & subscribed to stan on topic: ", subConf.Topic)

	return nil

}

func (rr *ReplicateReader) MsgC() <-chan *messages.SignedN3Message {
	return rr.out
}

func (rr *ReplicateReader) ErrC() <-chan error {
	return rr.client.ErrC
}

func (rr *ReplicateReader) Close() error {
	// note: as we use durable subscriptions DO NOT
	// unsubscribe, just close.
	return rr.client.sub.Close()
}
