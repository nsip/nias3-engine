// peer-feed-reader.go

package n3stan

import (
	"log"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
)

// Ensure PeerFeedReader implements n3.ReaderService interface
var _ n3.ReaderService = &PeerFeedReader{}

type PeerFeedReader struct {
	client *readerClient
	out    <-chan *messages.SignedN3Message
}

//
// Peer Feed Reader is a specialisation of the other readers
// as it require a non-durable subscription.
//
// the feed is always read copletely to ensure no gaps
// in data between peers.
//
func NewPeerFeedReader() n3.ReaderService {
	return &PeerFeedReader{}
}

func (pfr *PeerFeedReader) Subscribe(topic string, clientID string) error {

	var err error

	if clientID == "" {
		return errors.New("clientID cannot be empty string")
	}

	stanConf := DefaultStanConfig(clientID)
	pfr.client, err = NewReaderClient(stanConf)
	if err != nil {
		return errors.Wrap(err, "unable to create stan connection")
	}

	subConf := SubscriptionConfig{Topic: topic}
	err = pfr.client.nondurableSubscribe(subConf)
	if err != nil {
		return errors.Wrap(err, "unable to create subscription")
	}

	// pfr.out = make(chan *messages.SignedN3Message)
	pfr.out, err = createTransportReadAdaptor(pfr.client.MsgC, pfr.client.ErrC)
	if err != nil {
		return errors.Wrap(err, "unalbe to create feed stream read adaptor")
	}

	log.Println("PeerFeedReader successfully connected & subscribed to stan on topic: ", subConf.Topic)

	return nil

}

func (pfr *PeerFeedReader) MsgC() <-chan *messages.SignedN3Message {
	return pfr.out
}

func (pfr *PeerFeedReader) ErrC() <-chan error {
	return pfr.client.ErrC
}

func (pfr *PeerFeedReader) Close() error {
	// note: as we use durable subscriptions DO NOT
	// unsubscribe, just close.
	return pfr.client.sub.Close()
}
