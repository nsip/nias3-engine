// sync-request-reciprocator.go

package pipelines

import (
	"context"
	"log"
	"time"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/n3peer"
	"github.com/nsip/nias3-engine/storage/leveldb"
	"github.com/nsip/nias3-engine/stream/n3stan"
	"github.com/pkg/errors"
)

const DELAY_SECONDS = 2

func SyncRequestReciprocator(ctx context.Context, node *n3peer.Node, in <-chan *messages.SyncRequest) (
	<-chan error, error) {

	errc := make(chan error, 1)

	// ledger service to construct request digests
	ls, err := leveldb.NewLedgerService()
	if err != nil {
		return nil, err
	}

	// sync request writer!!!! NOT direct to node
	outboundSyncRequestWriter := n3stan.NewSyncRequestWriter()
	err = outboundSyncRequestWriter.Subscribe("sr.outbound", "sr-reciprocator")
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(errc)
		defer ls.Close()
		defer outboundSyncRequestWriter.Close()
		for srq := range in {

			log.Println("SyncRequestReciprocator:\tgot a message")
			inbound_sr := srq

			// send requests based on reasonable frequency
			time.Sleep(time.Second * DELAY_SECONDS)

			// create the reciprocal sync request
			digest, err := ls.CreateSyncDigest(inbound_sr.Context)
			if err != nil {
				errc <- errors.Wrap(err, "unable to create ledger digest:")
				continue
			}
			outbound_sr := &messages.SyncRequest{
				Context:      inbound_sr.Context,
				LedgerDigest: digest,
				SourceId:     inbound_sr.TargetId,
				TargetId:     inbound_sr.SourceId,
			}

			log.Printf("\t\toutbound sync-request:\n\n%s\n\n", outbound_sr)

			outboundSyncRequestWriter.MsgC() <- outbound_sr

			// select {
			// case <-ctx.Done():
			// 	return
			// case err := <-outboundSyncRequestWriter.ErrC():
			// 	errc <- err
			// }
		}
	}()

	return errc, nil

}
