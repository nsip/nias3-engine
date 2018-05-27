// sync.go

package n3

import (
	"fmt"
	"log"

	net "github.com/libp2p/go-libp2p-net"
	"github.com/pkg/errors"
)

// pattern: /protocol-name/request-or-response-message/version
const syncProtocol = "/n3/sync/0.0.1"

// SyncProtocol type
type SyncProtocol struct {
	node *Node // local host
}

func NewSyncProtocol(node *Node) *SyncProtocol {
	sp := &SyncProtocol{node: node}
	node.SetStreamHandler(syncProtocol, sp.handleStream)
	return sp
}

//
// handle connection between peers
//
func (sp *SyncProtocol) handleStream(s net.Stream) {

	log.Println("Got a new stream!")

	// wrap stream for protobuf exchange
	ws := WrapStream(s)

	// Create a buffer stream for non blocking read and write.
	// rw := bufio.NewReadWriter(ws.r, ws.w)

	go sp.ReadData(ws)
	go sp.WriteData(ws)

	// stream 's' will stay open until you close it (or the other side closes it).
}

//
// listen for transmitted blocks
// and add them to the local blockchains
//
func (sp *SyncProtocol) ReadData(ws *WrappedStream) {

	for {

		b, err := receiveBlock(ws)
		if err != nil {
			// log.Println("read-block error: ", err)
			break
		}

		log.Println("...got a message")

		if !b.Verify() {
			log.Println("recieved block failed verification %v", b)
			continue
		}

		bc := GetBlockchain(b.Data.Context, b.Author)
		bc.AddBlock(b.Data)

	}
}

//
// handles reading a block from the provided stream
//
func receiveBlock(ws *WrappedStream) (*Block, error) {
	var block Block
	err := ws.dec.Decode(&block)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read block from stream")
	}
	return &block, nil
}

//
// send block data to peers
//
func (sp *SyncProtocol) WriteData(ws *WrappedStream) {

	// go func() {
	// 	for {
	// 		time.Sleep(5 * time.Second)
	// 		mutex.Lock()
	// 		bytes, err := json.Marshal(Blockchain)
	// 		if err != nil {
	// 			log.Println(err)
	// 		}
	// 		mutex.Unlock()

	// 		mutex.Lock()
	// 		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	// 		rw.Flush()
	// 		mutex.Unlock()

	// 	}
	// }()

	//
	// change for channel reader
	//
	// stdReader := bufio.NewReader(os.Stdin)

	for {

		b := <-sp.node.BlockChan
		log.Println("...block received for writing ")
		err := sendBlock(b, ws)
		if err != nil {
			// log.Println("unable to write block to stream: ", err)
			break
		}

		// maybe also write tuple to feed

	}

}

//
// sendBlock encodes and writes a data-message to the outbound stream
//
func sendBlock(block *Block, ws *WrappedStream) error {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered error in send block", r)
		}
	}()

	err := ws.enc.Encode(block)
	// Because output is buffered with bufio, we need to flush!
	ws.w.Flush()
	return err
}
