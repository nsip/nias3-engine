// pipelined-ingest-writer.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/stream/n3stan"
)

// pipelined wrapper around the ingest-writer component.
//
// accepts SignedN3Mesages and writes to the ingest stream.
//
func PipelinedIngestWriter(ctx context.Context, in <-chan *messages.SignedN3Message) (<-chan error, error) {

	errc := make(chan error, 1)

	var ws n3.WriterService
	ws = n3stan.NewIngestWriter()
	err := ws.Subscribe("ingest", "pipe-igw")
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(errc)
		defer ws.Close()
		for signed_msg := range in {
			smsg := signed_msg
			select {
			case ws.MsgC() <- smsg:
			case err := <-ws.ErrC():
				errc <- err
			case <-ctx.Done():
				return
			}
		}
	}()
	return errc, nil

}
