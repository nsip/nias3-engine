// header-injector.go

package pipelines

import (
	"context"
	"log"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/n3crypto"
)

//
// header-injector is a reuseable pipleine utility to
// create a new block of meta-data for a message
//
func HeaderInjector(ctx context.Context, in <-chan *messages.SignedN3Message) (
	<-chan *messages.SignedN3Message, <-chan error, error) {

	out := make(chan *messages.SignedN3Message)
	errc := make(chan error, 1)
	cs, err := n3crypto.NewCryptoService()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer close(out)
		defer close(errc)
		for msg := range in {

			hdr_msg := msg

			// generate meta-data block
			msgData, err := cs.NewMessageData(false)
			if err != nil {
				log.Println("unable to generate message header meta-data: ", err)
				continue
			}
			hdr_msg.N3Message.MessageData = msgData

			// pass on
			select {
			case out <- hdr_msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, errc, nil
}
