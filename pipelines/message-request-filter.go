// message-request-filter.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
)

//
// MessageRequestFilter is a reusable pipeline component that
// ensures messages sent to remote peers are only those that they need
//
// the syncRequest from the remote peer is used to filter the stream of outgoing
// messages to ensure only the minimum required set is sent over the wire
//
func MessageRequestFilter(ctx context.Context, sReq *messages.SyncRequest,
	in <-chan *messages.SignedN3Message) (<-chan *messages.SignedN3Message, <-chan error, error) {

	out := make(chan *messages.SignedN3Message)
	errc := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errc)

		for msg := range in {

			fltr_msg := msg

			// check whether this is the required context
			if fltr_msg.N3Message.DataTuple.Context != sReq.Context {
				continue
			}
			// log.Println("contexts match")

			// check whether this message (user:context) is known to the requestor
			var lastSeq uint64
			key := fltr_msg.DigestKey()
			// log.Println("outbound msg digest key: ", key)
			lastSeq = sReq.LedgerDigest[key]
			// log.Println("lastSeq: ", lastSeq)
			// log.Println("outbound msg seq: ", fltr_msg.N3Message.Sequence)

			// pass on the message if last sequnce from requestor is < this message sequence
			// if last sequence is unknown (0) then pass on the message as requestor has not
			// seen messages from this user
			if lastSeq >= fltr_msg.N3Message.Sequence {
				// log.Println("...did not send msg")
				continue
			}

			// pass on
			select {
			case out <- fltr_msg:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, errc, nil

}
