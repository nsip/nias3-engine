// outbound-sync-request.go

package sync_request

import (
	"context"

	"github.com/nsip/nias3-engine/n3peer"
	pipes "github.com/nsip/nias3-engine/pipelines"
)

func OutboundSyncRequestPipeline(ctx context.Context, node *n3peer.Node) ([]<-chan error, error) {

	var errcList []<-chan error

	// read outbound requests from the stream
	srq_out, errc, err := pipes.PipelinedSyncRequestReader(ctx, pipes.OUTBOUND)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// for each request establish a sync protocol connection & send
	errc, err = pipes.SyncRequestSender(ctx, node, srq_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	return errcList, nil

}
