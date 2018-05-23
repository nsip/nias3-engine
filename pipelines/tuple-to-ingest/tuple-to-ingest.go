// tuple-to-ingest.go

package tuple_to_ingest

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
	pipes "github.com/nsip/nias3-engine/pipelines"
)

func TupleToIngestPipeline(ctx context.Context) (chan<- *messages.DataTuple, []<-chan error, error) {

	var errcList []<-chan error

	// provide an entry point to the pipelne that simply
	// receives data-tuples
	in := make(chan *messages.DataTuple)
	msg_out, errc, err := pipes.TupleReceiver(ctx, in)
	if err != nil {
		return nil, nil, err
	}
	errcList = append(errcList, errc)

	// add header meta-data
	hdr_msg_out, errc, err := pipes.HeaderInjector(ctx, msg_out)
	if err != nil {
		return nil, nil, err
	}
	errcList = append(errcList, errc)

	// create message signature
	signed_msg_out, errc, err := pipes.MessageSigner(ctx, hdr_msg_out)
	if err != nil {
		return nil, nil, err
	}
	errcList = append(errcList, errc)

	// write to the ingest stream
	errc, err = pipes.PipelinedIngestWriter(ctx, signed_msg_out)
	if err != nil {
		return nil, nil, err
	}
	errcList = append(errcList, errc)

	return in, errcList, nil
}
