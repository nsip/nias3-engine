// tuple-receiver.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
)

//
// tupleReceiver acts as first pipleine component to take
// tuples and wrap them as N3Messages.
//
// No real meta-data or hashing is created at this point
// that is added for real by the ingest-to-feed pipeline
//
func TupleReceiver(ctx context.Context, in <-chan *messages.DataTuple) (
	<-chan *messages.SignedN3Message, <-chan error, error) {

	out := make(chan *messages.SignedN3Message)
	errc := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errc)
		for tuple := range in {
			dt := tuple
			n3msg := &messages.SignedN3Message{}
			n3msg.N3Message = &messages.N3Message{DataTuple: dt}
			select {
			case out <- n3msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, errc, nil
}
