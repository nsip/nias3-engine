// inbound-sync-request.go

package sync_request

import (
	"context"

	"github.com/nsip/nias3-engine/n3peer"
	pipes "github.com/nsip/nias3-engine/pipelines"
)

func InboundSyncRequestPipeline(ctx context.Context, node *n3peer.Node) ([]<-chan error, error) {

	var errcList []<-chan error

	// read inbound requests from the stream
	srq_out, errc, err := pipes.PipelinedSyncRequestReader(ctx, pipes.INBOUND)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// send messages to remote peer
	recip_out, errc, err := pipes.SyncRequestFulfiller(ctx, node, srq_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// create a reciprocal sync request from the remote peer
	errc, err = pipes.SyncRequestReciprocator(ctx, node, recip_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	return errcList, nil

}
