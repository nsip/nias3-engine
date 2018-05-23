// message-authenticator.go

package pipelines

import (
	"context"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/n3crypto"
	"github.com/nsip/nias3-engine/stream/n3stan"
	"github.com/pkg/errors"
)

//
// message-authenticator is a reuseable pipleine utility to
// verify a signed message is untampered and from the
// sender who signed it
//
func MessageAuthenticator(ctx context.Context, in <-chan *messages.SignedN3Message) (
	<-chan *messages.SignedN3Message, <-chan error, error) {

	out := make(chan *messages.SignedN3Message)
	errc := make(chan error, 1)
	cs, err := n3crypto.NewCryptoService()
	if err != nil {
		return nil, nil, err
	}
	as, err := n3stan.NewAckService()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer close(out)
		defer close(errc)
		for msg := range in {

			auth_msg := msg
			msgID := auth_msg.N3Message.MessageData.Id

			valid, err := cs.AuthenticateMessage(auth_msg)
			if err != nil {
				as.Ack(msgID) //ensure problem message is not redelivered
				errc <- errors.Wrap(err, "error authenticating message from stream")
				continue
			}

			if !valid {
				as.Ack(msgID) //ensure problem message is not redelivered
				errc <- errors.Wrap(err, "auth: invalid message read from stream")
				continue
			}

			// pass on
			select {
			case out <- auth_msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, errc, nil
}
