// pipelined-replicate-reader.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/stream/n3stan"
)

//
// pipeline wrapper around replicate stream reader
//
// connects to the replicate stream and returns a channel from which to
// read messages
//
func PipelinedReplicateReader(ctx context.Context) (<-chan *messages.SignedN3Message, <-chan error, error) {

	// set up the reader for the ingest stream
	var rs n3.ReaderService
	rs = n3stan.NewReplicateReader()
	err := rs.Subscribe("replicate", "prr")
	if err != nil {
		return nil, nil, err
	}

	out := make(chan *messages.SignedN3Message)

	go func() {
		defer rs.Close()
		defer close(out)
		for {
			select {
			case msg := <-rs.MsgC():
				out <- msg
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, rs.ErrC(), nil

}
