// replicate-writer.go

package n3stan

import (
	"log"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
)

// Ensure ReplicateWriter implements n3.WriterService interface
var _ n3.WriterService = &ReplicateWriter{}

type ReplicateWriter struct {
	client *writerClient
	in     chan<- *messages.SignedN3Message
}

func NewReplicateWriter() n3.WriterService {
	return &ReplicateWriter{}
}

func (rw *ReplicateWriter) Subscribe(topic string, clientID string) error {

	var err error

	if clientID == "" {
		return errors.New("clientID cannot be empty string")
	}

	stanConf := DefaultStanConfig(clientID)
	rw.client, err = NewWriterClient(stanConf)
	if err != nil {
		return errors.Wrap(err, "unable to create stan connection")
	}

	subConf := SubscriptionConfig{Topic: topic, DurableName: stanConf.ClientId}
	err = rw.client.subscribe(subConf)
	if err != nil {
		return errors.Wrap(err, "unable to create subscription")
	}

	rw.in, err = createTransportWriteAdaptor(rw.client.MsgC, rw.client.ErrC)
	if err != nil {
		return errors.Wrap(err, "unable to create replicate stream write adaptor")
	}

	log.Println("ReplicateWriter successfully connected & subscribed to stan on topic: ", subConf.Topic)

	return nil

}

func (rw *ReplicateWriter) Close() error {
	//
	// publishing does not use a subscription
	// so this is a no-op
	//
	return nil
}

//
// Message channel to send domain mesages to be published
// will throw an error if Subscribe() has not been called
// before messages are sent to the channel
//
func (rw *ReplicateWriter) MsgC() chan<- *messages.SignedN3Message {
	return rw.in
}

func (rw *ReplicateWriter) ErrC() <-chan error {
	return rw.client.ErrC
}
