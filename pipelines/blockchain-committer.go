// blockchain-committer.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/storage/leveldb"
	"github.com/pkg/errors"
)

//
// blockchain-committer is a reuseable pipleine utility to
// insert the supplied message into the blockchain for this
// user/instance.
//
func BlockchainCommitter(ctx context.Context, in <-chan *messages.SignedN3Message) (
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
		defer ls.Close()
		for msg := range in {

			bc_msg := msg

			// commit the message
			err := ls.CommitNew(bc_msg.N3Message)
			if err != nil {
				errc <- errors.Wrap(err, "unable to commit message to ledger")
				continue
			}

			// pass on
			select {
			case out <- bc_msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, errc, nil
}
