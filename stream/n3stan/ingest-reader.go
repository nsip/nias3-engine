// ingest-reader.go

package n3stan

import (
	"log"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
)

// Ensure IngestReader implements n3.ReaderService interface
var _ n3.ReaderService = &IngestReader{}

type IngestReader struct {
	client *readerClient
	out    <-chan *messages.SignedN3Message
}

func NewIngestReader() n3.ReaderService {
	return &IngestReader{}
}

func (ir *IngestReader) Subscribe(topic string, clientID string) error {

	var err error

	if clientID == "" {
		return errors.New("clientID cannot be empty string")
	}

	stanConf := DefaultStanConfig(clientID)
	ir.client, err = NewReaderClient(stanConf)
	if err != nil {
		return errors.Wrap(err, "unable to create stan connection")
	}

	subConf := SubscriptionConfig{Topic: topic, DurableName: stanConf.ClientId}
	err = ir.client.subscribe(subConf)
	if err != nil {
		return errors.Wrap(err, "unable to create ingest subscription")
	}

	ir.out = make(chan *messages.SignedN3Message)
	ir.out, err = createOrderedTransportReadAdaptor(ir.client.MsgC, ir.client.ErrC)
	if err != nil {
		return errors.Wrap(err, "unable to create transport adaptor")
	}

	log.Println("IngestReader successfully connected & subscribed to stan on topic: ", subConf.Topic)

	return nil

}

func (ir *IngestReader) MsgC() <-chan *messages.SignedN3Message {
	return ir.out
}

func (ir *IngestReader) ErrC() <-chan error {
	return ir.client.ErrC
}

func (ir *IngestReader) Close() error {
	// note: as we use durable subscriptions DO NOT
	// unsubscribe, just close.
	err := ir.client.sub.Close()
	if err != nil {
		log.Println(err)
	}
	return err
	// ir.client.sub.Close()
}
