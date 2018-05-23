// sync-request-reader.go

package n3stan

import (
	"log"

	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
)

type SyncRequestReader struct {
	client *readerClient
	out    <-chan *messages.SyncRequest
}

func NewSyncRequestReader() *SyncRequestReader {
	return &SyncRequestReader{}
}

func (sr *SyncRequestReader) Subscribe(topic string, clientID string) error {

	var err error

	if clientID == "" {
		return errors.New("clientID cannot be empty string")
	}

	stanConf := DefaultStanConfig(clientID)
	sr.client, err = NewReaderClient(stanConf)
	if err != nil {
		return errors.Wrap(err, "unable to create stan connection")
	}

	subConf := SubscriptionConfig{Topic: topic, DurableName: stanConf.ClientId}
	err = sr.client.subscribe(subConf)
	if err != nil {
		return errors.Wrap(err, "unable to create ingest subscription")
	}

	sr.out = make(chan *messages.SyncRequest)
	sr.out, err = createSyncRequestReadAdaptor(sr.client.MsgC, sr.client.ErrC)
	if err != nil {
		return errors.Wrap(err, "unable to create transport adaptor")
	}

	log.Println("SyncRequestReader successfully connected & subscribed to stan on topic: ", subConf.Topic)

	return nil

}

func (sr *SyncRequestReader) MsgC() <-chan *messages.SyncRequest {
	return sr.out
}

func (sr *SyncRequestReader) ErrC() <-chan error {
	return sr.client.ErrC
}

func (sr *SyncRequestReader) Close() error {
	// note: as we use durable subscriptions DO NOT
	// unsubscribe, just close.
	err := sr.client.sub.Close()
	if err != nil {
		log.Println(err)
	}
	return err
	// ir.client.sub.Close()
}
