// p2p-message-writer.go

package pipelines

// import (
// 	"context"

// 	"github.com/nsip/nias3-engine/messages"
// 	"github.com/nsip/nias3-engine/n3peer"
// 	"github.com/pkg/errors"
// )

// //
// // P2PMessageWriter is a reusable pipeline component that
// // does the work of writing data sync messages to the
// // provided remote peer stream.
// //
// // this is a sink component that should be used at the end of a pipeline
// //
// func P2PMessageWriter(ctx context.Context, ws *n3peer.WrappedStream,
// 	in <-chan *messages.SignedN3Message) (<-chan error, error) {

// 	errc := make(chan error, 1)

// 	go func() {
// 		defer close(errc)

// 		for msg := range in {

// 			p2p_msg := msg

// 			err := n3peer.SendSyncMessage(p2p_msg, ws)

// 			// pass on
// 			select {
// 			case err != nil:
// 				errc <- errors.Wrap(err, "error sending message to remote peer")
// 			case <-ctx.Done():
// 				return
// 			default:
// 			}
// 		}
// 	}()

// 	return out, errc, nil

// }
