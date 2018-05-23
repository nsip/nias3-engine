// hexastore-committer.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/storage/leveldb"
	"github.com/pkg/errors"
)

//
// hexastore-committer is a sink pipeline component that has the task
// of adding data from a received message to the hexastore
//
func HexastoreCommitter(ctx context.Context, in <-chan *messages.SignedN3Message) (
	<-chan error, error) {

	errc := make(chan error, 1)

	hs, err := leveldb.NewHexastoreService()
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(errc)
		defer hs.Close()
		for msg := range in {

			commit_msg := msg

			// commit the data
			err := hs.Commit(commit_msg.N3Message)
			if err != nil {
				errc <- errors.Wrap(err, "unable to commit tuple to hexastore")
				continue
			}

			// wait for close
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	return errc, nil

}
