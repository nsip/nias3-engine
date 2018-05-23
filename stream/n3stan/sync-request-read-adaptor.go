// sync-request-read-adaptor.go

package n3stan

import (
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats-streaming"
	"github.com/nsip/nias3-engine/messages"
)

//
// Read adaptor assumes it will receive transport messages (stan.Msg) from the
// streaming transport, and extract & pass on the domain messsage (SyncRequest in this case)
//
//
// Messages are acked here, because even if unexpected/unsuitable messages are read from the
// ingest feed we want them to not be redelivered, and re-ordering through re-delivery
// is not required for the sync request feed.
//
type SyncRequestReadAdaptor struct {
}

// params:
// in - channel which receives messages from the stan topic - typically the MsgC of the readerClient
// errc - channel to pass out significant errors - typically the ErrC of the readerClient
//
// returns - a channel to listen for converted domain messages
//
func createSyncRequestReadAdaptor(in <-chan *stan.Msg, errc chan<- error) (chan *messages.SyncRequest, error) {

	out := make(chan *messages.SyncRequest)
	go func() {
		defer close(out)
		for msg := range in {
			// extract data
			smsg := &messages.SyncRequest{}
			if err := proto.Unmarshal(msg.Data, smsg); err != nil {
				log.Println("unexpected message from sync-request feed, message discarded.")
				msg.Ack() // ack even if error as we don't want bad messages re-delivered
				continue
			}
			out <- smsg
			msg.Ack()
		}
	}()
	return out, nil

}
