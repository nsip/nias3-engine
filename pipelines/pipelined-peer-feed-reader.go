// pipelined-peer-feed-reader.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/stream/n3stan"
)

//
// pipeline wrapper around a specialised feed reader
//
// connects to the feed stream and returns a channel from which to
// read messages
//
// difference with this reader is it always traverses the full feed
// to ensure there are no gaps in data
//
func PipelinedPeerFeedReader(ctx context.Context) (<-chan *messages.SignedN3Message, <-chan error, error) {

	// set up the reader for the ingest stream
	var rs n3.ReaderService
	rs = n3stan.NewPeerFeedReader()
	err := rs.Subscribe("feed", "p2pfr")
	if err != nil {
		return nil, nil, err
	}

	out := make(chan *messages.SignedN3Message)

	go func() {
		defer rs.Close()
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
