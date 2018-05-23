// ingest-writer.go

package n3stan

import (
	"log"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
)

// Ensure IngestWriter implements n3.WriterService interface
var _ n3.WriterService = &IngestWriter{}

type IngestWriter struct {
	client *writerClient
	in     chan<- *messages.SignedN3Message
}

func NewIngestWriter() n3.WriterService {
	return &IngestWriter{}
}

func (iw *IngestWriter) Subscribe(topic string, clientID string) error {

	var err error

	if clientID == "" {
		return errors.New("clientID cannot be empty string")
	}

	stanConf := DefaultStanConfig(clientID)
	iw.client, err = NewWriterClient(stanConf)
	if err != nil {
		return errors.Wrap(err, "unable to create stan connection")
	}

	subConf := SubscriptionConfig{Topic: topic, DurableName: stanConf.ClientId}
	err = iw.client.subscribe(subConf)
	if err != nil {
		return errors.Wrap(err, "unable to create subscription")
	}

	iw.in, err = createTransportWriteAdaptor(iw.client.MsgC, iw.client.ErrC)
	if err != nil {
		return errors.Wrap(err, "unable to create ingest stream write adaptor")
	}

	log.Println("IngestWriter successfully connected & subscribed to stan on topic: ", subConf.Topic)

	return nil

}

func (iw *IngestWriter) Close() error {
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
func (iw *IngestWriter) MsgC() chan<- *messages.SignedN3Message {
	return iw.in
}

func (iw *IngestWriter) ErrC() <-chan error {
	return iw.client.ErrC
}
