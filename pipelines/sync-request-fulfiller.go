// sync-request-fulfiller.go

package pipelines

import (
	"context"
	"log"
	"sync"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/n3peer"
	"github.com/pkg/errors"
)

//
// sync-request-fulfiller is a reuseable pipleine utility to
// send data messages to a remote peer based on a received
// synchronisation request, which allows the data sent to be
// filtered so that only new data (to the remote peer) is sent.
//
func SyncRequestFulfiller(ctx context.Context, node *n3peer.Node, in <-chan *messages.SyncRequest) (
	<-chan *messages.SyncRequest, <-chan error, error) {

	out := make(chan *messages.SyncRequest)
	errc := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errc)
		for sr := range in {

			log.Println("SyncRequestFulfiller:\t got sync request")

			sync_req := sr

			var wg sync.WaitGroup

			wg.Add(1)

			go func() {
				// 		get a wrapped stream from node
				ws, err := node.GetWrappedSyncResponseStream(sync_req.SourceId)
				defer node.CloseSyncResponseStream(ws)

				// 		open the feed stream
				feed_msgs, f2p_errc, err := feedToPeerPipeline(ctx, sync_req)
				if err != nil {
					errc <- errors.Wrap(err, "fulfiller unable to connect to feed")
					// continue
				}
				// allow any feed-reading errors to be reported
				go MonitorPipeline("feed-to-peer", f2p_errc...)

				//  read feed messages unitl up to date
				for msg := range feed_msgs {
					smsg := msg
					// send the message to remote peer
					err := node.SendSyncResponseMessage(smsg, ws)
					if err != nil {
						log.Println("SyncRequestFulfller: send error: ", err)
					}
					log.Println("SyncRequestFulfiller:\tmessage sent")
					// select {
					// // pass out any send errors
					// case errc <- err:
					// case <-ctx.Done():
					// 	return
					// }
				}
				wg.Done()
			}()

			wg.Wait()

			// once processing complete pass on the
			// fulfilled sync request, & listen for shutdowns
			// select {
			// case out <- sync_req:
			// 	log.Println("SyncRequestFulfiller:\tall messages sent reciprocating...")
			// case <-ctx.Done():
			// 	return
			// }
			log.Println("SyncRequestFulfiller:\tpassing on sync request")
			out <- sync_req
		}
	}()

	return out, errc, nil

}

//
// pipeline reads from Feed, filters messages for sending to a
// remote peer, and returns a channel to listen for those
// messages, typically so they can be written to a remote
// stream by the p2p node.
//
func feedToPeerPipeline(ctx context.Context, sReq *messages.SyncRequest) (
	<-chan *messages.SignedN3Message, []<-chan error, error) {

	var errcList []<-chan error

	// attach & read from the Feed stream
	msg_out, errc, err := PipelinedPeerFeedReader(ctx)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	errcList = append(errcList, errc)

	// only send feed messages if required
	feed_msg_out, errc, err := MessageRequestFilter(ctx, sReq, msg_out)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	errcList = append(errcList, errc)

	// close pipeline if no new messages received within a deadline
	out, errc, err := P2PFeedCloser(ctx, feed_msg_out)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	errcList = append(errcList, errc)

	return out, errcList, nil
}
