// feed-reader.go

package n3stan

import (
	"log"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
)

// Ensure FeedReader implements n3.ReaderService interface
var _ n3.ReaderService = &FeedReader{}

type FeedReader struct {
	client *readerClient
	out    <-chan *messages.SignedN3Message
}

func NewFeedReader() n3.ReaderService {
	return &FeedReader{}
}

func (fr *FeedReader) Subscribe(topic string, clientID string) error {

	var err error

	if clientID == "" {
		return errors.New("clientID cannot be empty string")
	}

	stanConf := DefaultStanConfig(clientID)
	fr.client, err = NewReaderClient(stanConf)
	if err != nil {
		return errors.Wrap(err, "unable to create stan connection")
	}

	subConf := SubscriptionConfig{Topic: topic, DurableName: "durable" + stanConf.ClientId}
	err = fr.client.subscribe(subConf)
	if err != nil {
		return errors.Wrap(err, "unable to create subscription")
	}

	fr.out = make(chan *messages.SignedN3Message)
	fr.out, err = createOrderedTransportReadAdaptor(fr.client.MsgC, fr.client.ErrC)
	if err != nil {
		return errors.Wrap(err, "unalbe to create feed stream read adaptor")
	}

	log.Println("FeedReader successfully connected & subscribed to stan on topic: ", subConf.Topic)

	return nil

}

func (fr *FeedReader) MsgC() <-chan *messages.SignedN3Message {
	return fr.out
}

func (fr *FeedReader) ErrC() <-chan error {
	return fr.client.ErrC
}

func (fr *FeedReader) Close() error {
	// note: as we use durable subscriptions DO NOT
	// unsubscribe, just close.
	log.Println("feed-reader closing")
	return fr.client.sub.Close()
}
