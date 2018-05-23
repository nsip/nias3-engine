// boundary-assigner.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/storage/leveldb"
	"github.com/pkg/errors"
)

//
// sequnce-assigner is a reuseable pipleine utility to
// determine the next available boundary/version number for
// the data tuple contained in the message
//
func BoundaryAssigner(ctx context.Context, in <-chan *messages.SignedN3Message) (
	<-chan *messages.SignedN3Message, <-chan error, error) {

	out := make(chan *messages.SignedN3Message)
	errc := make(chan error, 1)
	bs, err := leveldb.NewBoundaryService()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer close(out)
		defer close(errc)
		for msg := range in {

			bnd_msg := msg

			// set next seq. no
			err := bs.AssignBoundary(bnd_msg.N3Message)
			if err != nil {
				errc <- errors.Wrap(err, "unable to assign message tuple-data boundary")
				continue
			}

			// pass on
			select {
			case out <- bnd_msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, errc, nil
}
