// feed-to-hex.go

package feed_to_hex

import (
	"context"

	pipes "github.com/nsip/nias3-engine/pipelines"
)

func FeedToHexPipeline(ctx context.Context) ([]<-chan error, error) {

	var errcList []<-chan error

	// attach & read from the Feed stream
	msg_out, errc, err := pipes.PipelinedFeedReader(ctx)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// validate feed messages
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

	// resolve any data boundary conflicts
	cr_msg_out, errc, err := pipes.ConflictResolver(ctx, ack_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	// write data to the hexastore
	errc, err = pipes.HexastoreCommitter(ctx, cr_msg_out)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	return errcList, nil
}
