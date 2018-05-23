// pipelined-replicate-writer.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/stream/n3stan"
)

// pipelined wrapper around the replicate-writer component.
//
// accepts SignedN3Mesages and writes to the replicate stream.
//
func PipelinedReplicateWriter(ctx context.Context, in <-chan *messages.SignedN3Message) (<-chan error, error) {

	errc := make(chan error, 1)

	var ws n3.WriterService
	ws = n3stan.NewReplicateWriter()
	err := ws.Subscribe("replicate", "pipe-rplw")
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(errc)
		for signed_msg := range in {
			smsg := signed_msg
			select {
			case ws.MsgC() <- smsg:
				// log.Print(smsg.Print())
			case err := <-ws.ErrC():
				errc <- err
			case <-ctx.Done():
				return
			}
		}
	}()
	return errc, nil

}
