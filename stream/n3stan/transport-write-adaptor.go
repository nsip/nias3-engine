// transport-write-adaptor.go

package n3stan

import (
	"github.com/gogo/protobuf/proto"
	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
)

//
// Write adaptor assumes it will receive domain messages (SignedN3Message) from the
// application, and create transport-layer message payloads ([]byte - protobufs in this case)
//
//
type TransportWriteAdaptor struct {
}

// params:
// out - the channel of the writerClient to send messages to
// erc - a channel to pass significant errors to; typically the ErrC from the writerClient
//
// returns - a channel that can receive doamin messages for publishing
//
func createTransportWriteAdaptor(out chan<- []byte, errc chan<- error) (chan<- *messages.SignedN3Message, error) {

	in := make(chan *messages.SignedN3Message)
	go func() {
		defer close(in)

		for msg := range in {
			// log.Println("write adaptor received message")
			byteMsg, err := proto.Marshal(msg)
			if err != nil {
				errc <- errors.Wrap(err, "unable to marshal signed message")
				continue
			}

			select {
			case out <- byteMsg:
			}
		}
	}()
	return in, nil

}
