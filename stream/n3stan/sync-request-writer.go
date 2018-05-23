// sync-request-writer.go

package n3stan

import (
	"log"

	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
)

type SyncRequestWriter struct {
	client *writerClient
	in     chan<- *messages.SyncRequest
}

func NewSyncRequestWriter() *SyncRequestWriter {
	return &SyncRequestWriter{}
}

func (srw *SyncRequestWriter) Subscribe(topic string, clientID string) error {

	var err error

	if clientID == "" {
		return errors.New("clientID cannot be empty string")
	}

	stanConf := DefaultStanConfig(clientID)
	srw.client, err = NewWriterClient(stanConf)
	if err != nil {
		return errors.Wrap(err, "unable to create stan connection")
	}

	subConf := SubscriptionConfig{Topic: topic, DurableName: stanConf.ClientId}
	err = srw.client.subscribe(subConf)
	if err != nil {
		return errors.Wrap(err, "unable to create subscription")
	}

	srw.in, err = createSyncRequestWriteAdaptor(srw.client.MsgC, srw.client.ErrC)
	if err != nil {
		return errors.Wrap(err, "unable to create ingest stream write adaptor")
	}

	log.Println("SyncRequestWriter successfully connected & subscribed to stan on topic: ", subConf.Topic)

	return nil

}

func (srw *SyncRequestWriter) Close() error {
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
func (srw *SyncRequestWriter) MsgC() chan<- *messages.SyncRequest {
	return srw.in
}

func (srw *SyncRequestWriter) ErrC() <-chan error {
	return srw.client.ErrC
}
