// message-signer.go

package pipelines

import (
	"context"
	"log"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/n3crypto"
)

//
// message-signer is a reuseable pipleine utility to
// add a digital signature to a N3Message, accepts
// SignedN3Messages on its input channel and outputs
// re-signed SignedN3Messages.
//
func MessageSigner(ctx context.Context, in <-chan *messages.SignedN3Message) (
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

			raw_msg := msg

			// sign the complete message
			signed_msg, err := cs.SignMessage(raw_msg.N3Message)
			if err != nil {
				log.Println("unable to create signed message: ", err)
				continue
			}

			// pass on
			select {
			case out <- signed_msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, errc, nil
}
