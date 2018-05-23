// feed-to-peer.go

package feed_to_peer

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
	pipes "github.com/nsip/nias3-engine/pipelines"
)

//
// pipeline reads from Feed, filters messages for sending to a
// remote peer, and returns a channel to listen for those
// messages, typically so they can be written to a remote
// stream by the p2p node.
//
func FeedToPeerPipeline(ctx context.Context, sReq *messages.SyncRequest) (
	<-chan *messages.SignedN3Message, []<-chan error, error) {

	var errcList []<-chan error

	// attach & read from the Feed stream
	msg_out, errc, err := pipes.PipelinedPeerFeedReader(ctx)
	if err != nil {
		return nil, nil, err
	}
	errcList = append(errcList, errc)

	// only send feed messages if required
	feed_msg_out, errc, err := pipes.MessageRequestFilter(ctx, sReq, msg_out)
	if err != nil {
		return nil, nil, err
	}
	errcList = append(errcList, errc)

	// close pipeline if no new messages received within a deadline
	out, errc, err := pipes.P2PFeedCloser(ctx, feed_msg_out)
	if err != nil {
		return nil, nil, err
	}
	errcList = append(errcList, errc)

	return out, errcList, nil
}
