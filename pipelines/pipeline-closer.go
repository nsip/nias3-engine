// pipeline-closer.go

package pipelines

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/nsip/nias3-engine/messages"
)

//
// PipelineCloser
//
// Streams from nats have long-lived handlers, to make sure
// delivery to a remote peer is a discrete operation so that connections
// are only open for the minimum time required this component
// simply passes on messages,  but will invoke the cancellation of the
// pipeline when no new messages have been received for a period of time
//
// this shouuld always be the last sink compnent of a pipeline to
// ensure all previous processors are closed cleanly
//
func PipelineCloser(ctx context.Context, in <-chan *messages.SignedN3Message) (<-chan error, error) {

	// out := make(chan *messages.SignedN3Message)
	errc := make(chan error, 1)

	// _, cancelFunc := context.WithCancel(ctx)

	go func() {
		defer close(errc)
		// defer close(out)

		for {
			select {
			case <-in:
				// out <- msg
				log.Println("saw a message")
			case <-time.After(2000 * time.Millisecond):
				log.Println("...no more messages available")
				errc <- errors.New("pipeline-shutdown")
				// cancelFunc()
				return
			case <-ctx.Done():
				return
			}
		}

	}()

	return errc, nil

}
