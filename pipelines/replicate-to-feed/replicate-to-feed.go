// replicate-to-feed.go

package replicate_to_feed

import (
	"context"

	pipes "github.com/nsip/nias3-engine/pipelines"
)

//
//
//
func ReplicateToFeedPipeline(ctx context.Context) ([]<-chan error, error) {

	var errcList []<-chan error

	msg_out, errc, err := pipes.PipelinedReplicateReader(ctx)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// validate received messages
	valid_msg_out, errc, err := pipes.MessageAuthenticator(ctx, msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// enforce blockchain ordering of received messages
	ordered_msg_out, errc, err := pipes.BlockchainAppender(ctx, valid_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// ack any messages that got this far to prevent redelivery
	ack_msg_out, errc, err := pipes.MessageAcknowledger(ctx, ordered_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// assign sequence
	seq_msg_out, errc, err := pipes.SequenceAssigner(ctx, ack_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// write to feed stream
	errc, err = pipes.PipelinedFeedWriter(ctx, seq_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	return errcList, nil
}
