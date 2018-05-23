// ack-service.go

package n3stan

import (
	"sync"

	"github.com/nats-io/go-nats-streaming"
)

var msgRefs = make(map[string]*stan.Msg)
var mutex = &sync.Mutex{}

//
// simple mutex-protected servie that maintains a list
// of stan messages awaiting acknowledgement
// allows ack to be deferred in a pipeline until
// appropriate to the business logic.
//
type AckService struct {
	refs map[string]*stan.Msg
	mtx  *sync.Mutex
}

func NewAckService() (*AckService, error) {

	as := &AckService{refs: msgRefs, mtx: mutex}
	return as, nil

}

//
// add a message to the ackservice store
// params:
// id: message id from the MessageData header block
// msgRef: pointer to the original stan msg
//
func (as *AckService) Put(id string, msgRef *stan.Msg) {

	as.mtx.Lock()
	as.refs[id] = msgRef
	as.mtx.Unlock()

}

//
// acknowledge a message so it will not be re-delivered
//
func (as *AckService) Ack(id string) {

	as.mtx.Lock()
	msg, ok := as.refs[id]
	if ok {
		msg.Ack()
		delete(as.refs, id)
		// log.Println("message acked: ", id)
	}

	as.mtx.Unlock()

}
