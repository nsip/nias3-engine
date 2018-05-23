// peer-to-replicate.go

package peer_to_replicate

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
	pipes "github.com/nsip/nias3-engine/pipelines"
)

func PeerToReplicatePipeline(ctx context.Context) (chan<- *messages.SignedN3Message, []<-chan error, error) {

	var errcList []<-chan error

	// provide input channel for messages received from a remote peer
	in := make(chan *messages.SignedN3Message)

	// write to the ingest stream
	errc, err := pipes.PipelinedReplicateWriter(ctx, in)
	if err != nil {
		return nil, nil, err
	}
	errcList = append(errcList, errc)

	return in, errcList, nil
}
