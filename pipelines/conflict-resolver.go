// conflict-resolver.go

package pipelines

import (
	"context"
	"strings"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/storage/leveldb"
	"github.com/pkg/errors"
)

//
// conflict-resolver is a reuseable pipleine utility to
// enforce consistency of data at all nodes.
//
//
// algorithm is basically the same as the default couchdb/gundb
// auto conflict resolution:
//
// changes above the current boundary are allowed through, the enforced ordering from
// the blockchain ensures data cannot be supplied with far-future versions unless all
// previous versions have been seen.
//
// changes that fall below the current boundary are rejected as they cannot be
// causally consistent.
//
// where the current boundary and supplied boundary match, then data is committed if it is
// lexicographically greater
// key thing is that this behaviour is consistent on all nodes, so the data will be
// consistent in all places after a round of replication.
//
func ConflictResolver(ctx context.Context, in <-chan *messages.SignedN3Message) (
	<-chan *messages.SignedN3Message, <-chan error, error) {

	out := make(chan *messages.SignedN3Message)
	errc := make(chan error, 1)
	bs, err := leveldb.NewBoundaryService()
	if err != nil {
		return nil, nil, err
	}
	hs, err := leveldb.NewHexastoreService()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer close(out)
		defer close(errc)
		defer bs.Close()
		defer hs.Close()
		for msg := range in {

			cr_msg := msg
			msg_version := cr_msg.N3Message.DataTuple.Version

			// get current boundary number for this tuple
			last, err := bs.GetLast(cr_msg.N3Message)
			if err != nil {
				errc <- errors.Wrap(err, "unable to retrieve c/r bouundary value for tuple")
				continue
			}

			// ignore messages below baoundary
			if last > msg_version {
				continue
			}

			// if conflict pass to resolver
			if msg_version == last {
				commit, err := commitConflictingVersion(cr_msg.N3Message.DataTuple, hs)
				if err != nil {
					errc <- errors.Wrap(err, "unable to resolve commit conflict")
				}
				if !commit {
					continue
				}
			}

			// if message above boundary, pass on (no-op)
			// note: this relies on messages having been received in order
			// so that messages with a high future boundary don't cause gaps.

			// pass on
			select {
			case out <- cr_msg:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, errc, nil

}

//
// this function checks the content of provided message against the
// current version in the hexastore, and allows changes through if they are
// greater than the current version.
//
func commitConflictingVersion(dt *messages.DataTuple, hs n3.HexastoreService) (bool, error) {

	current, err := hs.GetCurrentValue(dt)
	if err != nil {
		return false, err
	}

	// trim all strings
	currVal := strings.Trim(current, " ")
	tupleVal := strings.Trim(dt.Object, " ")

	// no entry found so we can commit the supplied tuple
	if currVal == "" {
		return true, nil
	}

	// if entry is the same no need to commit
	if currVal == tupleVal {
		return false, nil
	}

	// see if the existing tuple is lex. greater than in msgs
	if currVal > tupleVal {
		return false, nil
	}

	// otherwise can commit
	return true, nil
}
