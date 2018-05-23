// sync-request-sender.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/n3peer"
	"github.com/pkg/errors"
)

//
// SyncRequestSender is responsible for
// passing on synchronisation (replicaiton) requests
// to other peers
//
//
func SyncRequestSender(ctx context.Context, node *n3peer.Node, in <-chan *messages.SyncRequest) (
	<-chan error, error) {

	errc := make(chan error, 1)

	go func() {
		defer close(errc)
		for srq := range in {
			// log.Println("SyncRequestSender:\tgot message")
			request := srq
			err := node.SendSyncRequest(request)
			if err != nil {
				errc <- errors.Wrap(err, "unable to send sync-req to remote peer:")
				continue
			}
			// log.Println("SyncRequestSender:\tSent message to remote peer")
		}
		// select {
		// case <-ctx.Done():
		// 	return
		// }
	}()

	return errc, nil

}
