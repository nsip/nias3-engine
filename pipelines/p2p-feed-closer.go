// p2p-feed-closer.go

package pipelines

import (
	"context"
	"log"
	"time"

	"github.com/nsip/nias3-engine/messages"
)

//
// P2PFeedCloser
//
// Streams from nats have long-lived handlers, to make sure
// delivery to a remote peer is a discrete operation so that connections
// are only open for the minimum time required this component
// simply passes on messages,  but will invoke the cancellation of the
// pipeline when no new messages have been received for a period of time
//
func P2PFeedCloser(ctx context.Context, in <-chan *messages.SignedN3Message) (
	<-chan *messages.SignedN3Message, <-chan error, error) {

	out := make(chan *messages.SignedN3Message)
	errc := make(chan error, 1)

	go func() {
		defer close(errc)
		defer close(out)

		for {
			select {
			case msg := <-in:
				out <- msg
			case <-time.After(500 * time.Millisecond):
				log.Println("no more messages from feed available")
				// errc <- errors.New("feed-closed")
				return
			case <-ctx.Done():
				return

			}
		}

	}()

	return out, errc, nil

}
