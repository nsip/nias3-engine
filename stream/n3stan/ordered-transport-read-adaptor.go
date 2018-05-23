// ordered-transport-read-adaptor.go

package n3stan

import (
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats-streaming"
	"github.com/nsip/nias3-engine/messages"
)

//
// Specialisation of the transport-read-adaptor.
//
// This adaptor is used to maintian a reference to the original stan msg
// in the Acknowledgement service, before passing on the message
// contents as a domain-level message.
//
// references to the original msg are stored using the msg id of the
// domain-level message.
//
// this allows downstream compoennts to delay acknowledgement of the
// message based on business logic, most downstream checks will result in
// the message being acknowledged, the primary one that won't is when
// a message is received out of order for the underlying block chain.
//
// by not acknowledging these messages they are re-delivered over time from the server
// meaning that eventually the correct odrder of messages will be re-stablished.
//
type OrderedTransportReadAdaptor struct {
}

// params:
// in - channel which receives messages from the stan topic - typically the MsgC of the readerClient
// errc - channel to pass out significant errors - typically the ErrC of the readerClient
//
// returns - a channel to listen for converted domain messages
//
func createOrderedTransportReadAdaptor(in <-chan *stan.Msg, errc chan<- error) (chan *messages.SignedN3Message, error) {

	out := make(chan *messages.SignedN3Message)
	ackSrv, err := NewAckService()
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(out)

		for msg := range in {

			keepMsg := msg

			// transform transport-level []bytes into domain message
			smsg := &messages.SignedN3Message{}
			if err := proto.Unmarshal(keepMsg.Data, smsg); err != nil {
				log.Println("unexpected message from replication stream, message discarded.")
				keepMsg.Ack() // ack on error as we don't want bad messages re-delivered
				continue
			}

			// store msg reference in the ack service
			id := smsg.N3Message.MessageData.Id
			ackSrv.Put(id, keepMsg)

			select {
			case out <- smsg:
			}
		}
	}()
	return out, nil

}
