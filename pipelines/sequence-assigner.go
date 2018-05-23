// sequence-assigner.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/storage/leveldb"
	"github.com/pkg/errors"
)

//
// sequnce-assigner is a reuseable pipleine utility to
// determine the next available sequence number for
// messages that are to be added to the feed
//
func SequenceAssigner(ctx context.Context, in <-chan *messages.SignedN3Message) (
	<-chan *messages.SignedN3Message, <-chan error, error) {

	out := make(chan *messages.SignedN3Message)
	errc := make(chan error, 1)
	ls, err := leveldb.NewLedgerService()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer close(out)
		defer close(errc)
		for msg := range in {

			seq_msg := msg

			// set next seq. no
			err := ls.AssignSequence(seq_msg.N3Message)
			if err != nil {
				errc <- errors.Wrap(err, "unable to assign message sequence")
				continue
			}

			// pass on
			select {
			case out <- seq_msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, errc, nil
}
