// feed-writer.go

package n3stan

import (
	"log"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
)

// Ensure FeedtWriter implements n3.WriterService interface
var _ n3.WriterService = &FeedWriter{}

type FeedWriter struct {
	client *writerClient
	in     chan<- *messages.SignedN3Message
}

func NewFeedWriter() n3.WriterService {
	return &FeedWriter{}
}

func (fw *FeedWriter) Subscribe(topic string, clientID string) error {

	var err error

	if clientID == "" {
		return errors.New("clientID cannot be empty string")
	}

	stanConf := DefaultStanConfig(clientID)
	fw.client, err = NewWriterClient(stanConf)
	if err != nil {
		return errors.Wrap(err, "unable to create stan connection")
	}

	subConf := SubscriptionConfig{Topic: topic, DurableName: stanConf.ClientId}
	err = fw.client.subscribe(subConf)
	if err != nil {
		return errors.Wrap(err, "unable to create subscription")
	}

	fw.in, err = createTransportWriteAdaptor(fw.client.MsgC, fw.client.ErrC)
	if err != nil {
		return errors.Wrap(err, "unable to create feed stream write adaptor")
	}

	log.Println("FeedWriter successfully connected & subscribed to stan on topic: ", subConf.Topic)

	return nil

}

func (fw *FeedWriter) Close() error {
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
func (fw *FeedWriter) MsgC() chan<- *messages.SignedN3Message {
	return fw.in
}

func (fw *FeedWriter) ErrC() <-chan error {
	return fw.client.ErrC
}
