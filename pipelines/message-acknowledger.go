// message-acknowledger.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/stream/n3stan"
)

//
// message-acknowledger is a reuseable pipleine utility to
// ack a message from the streaming transport layer
// at any point in a pipleine & therefore prevent unwanted re-delivery
//
func MessageAcknowledger(ctx context.Context, in <-chan *messages.SignedN3Message) (
	<-chan *messages.SignedN3Message, <-chan error, error) {

	out := make(chan *messages.SignedN3Message)
	errc := make(chan error, 1)
	as, err := n3stan.NewAckService()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer close(out)
		defer close(errc)
		for msg := range in {

			ack_msg := msg
			msgID := ack_msg.N3Message.MessageData.Id
			as.Ack(msgID) //ensure message is not redelivered

			// pass on
			select {
			case out <- ack_msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, errc, nil

}
