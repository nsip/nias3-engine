// pipelined-sync-request-reader.go

package pipelines

import (
	"context"
	"log"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/stream/n3stan"
)

const (
	INBOUND = iota
	OUTBOUND
)

//
// pipeline wrapper around sync request reader
//
// connects to the sync request stream and returns a channel from which to
// read messages
//
func PipelinedSyncRequestReader(ctx context.Context, direction int) (
	<-chan *messages.SyncRequest, <-chan error, error) {

	// set up the reader for the ingest stream
	srrs := n3stan.NewSyncRequestReader()
	var subErr error
	if direction == OUTBOUND {
		subErr = srrs.Subscribe("sr.outbound", "srq-ordr")
	} else {
		subErr = srrs.Subscribe("sr.inbound", "srq-irdr")
	}
	if subErr != nil {
		return nil, nil, subErr
	}

	out := make(chan *messages.SyncRequest)

	go func() {
		defer srrs.Close()
		defer close(out)
		for {
			select {
			case request := <-srrs.MsgC():
				log.Println("PipelinedSyncRequestReader:\tgot request - dir:", direction)
				out <- request
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, srrs.ErrC(), nil

}
