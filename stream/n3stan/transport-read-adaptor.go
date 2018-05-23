// transport-read-adaptor.go

package n3stan

import (
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats-streaming"
	"github.com/nsip/nias3-engine/messages"
)

//
// Read adaptor assumes it will receive transport messages (stan.Msg) from the
// streaming transport, and extract & pass on the domain messsage (SignedN3Message)
//
//
// Messages are acked here, because even if unexpected/unsuitable messages are read from the
// ingest feed we want them to not be redelivered, and re-ordering through re-delivery
// is not required for the ingest feed.
//
type TransportReadAdaptor struct {
}

// params:
// in - channel which receives messages from the stan topic - typically the MsgC of the readerClient
// errc - channel to pass out significant errors - typically the ErrC of the readerClient
//
// returns - a channel to listen for converted domain messages
//
func createTransportReadAdaptor(in <-chan *stan.Msg, errc chan<- error) (chan *messages.SignedN3Message, error) {

	out := make(chan *messages.SignedN3Message)
	go func() {
		defer close(out)
		for msg := range in {
			// extract data
			smsg := &messages.SignedN3Message{}
			if err := proto.Unmarshal(msg.Data, smsg); err != nil {
				log.Println("unexpected message from ingest feed, message discarded.")
				msg.Ack() // ack even if error as we don't want bad messages re-delivered
				continue
			}
			out <- smsg
			msg.Ack()
		}
	}()
	return out, nil

}
