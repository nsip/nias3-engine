// ingest-to-feed.go

package ingest_to_feed

import (
	"context"

	pipes "github.com/nsip/nias3-engine/pipelines"
)

func IngestToFeedPipeline(ctx context.Context) ([]<-chan error, error) {

	var errcList []<-chan error

	msg_out, errc, err := pipes.PipelinedIngestReader(ctx)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// validate ingested messages
	valid_msg_out, errc, err := pipes.MessageAuthenticator(ctx, msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// this pipeline requires no redelivery, so ack the messages at transport layer
	ack_msg_out, errc, err := pipes.MessageAcknowledger(ctx, valid_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// assign seq. no & store
	seq_msg_out, errc, err := pipes.SequenceAssigner(ctx, ack_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// assign data tuple boundary & store
	bnd_msg_out, errc, err := pipes.BoundaryAssigner(ctx, seq_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// commit to ledger
	bc_msg_out, errc, err := pipes.BlockchainCommitter(ctx, bnd_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// re-sign message
	signed_msg_out, errc, err := pipes.MessageSigner(ctx, bc_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// write to feed stream
	errc, err = pipes.PipelinedFeedWriter(ctx, signed_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	return errcList, nil

}
