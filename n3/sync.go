// sync.go

package n3

import (
	"fmt"
	"log"
	"sync"

	net "github.com/libp2p/go-libp2p-net"
	"github.com/nats-io/go-nats-streaming"
	"github.com/nats-io/nuid"
	"github.com/pkg/errors"
)

// pattern: /protocol-name/request-or-response-message/version
const syncProtocol = "/n3/sync/0.0.1"

var mutex = &sync.Mutex{}

// SyncProtocol type
type SyncProtocol struct {
	node *Node // local host
}

func NewSyncProtocol(node *Node) *SyncProtocol {
	sp := &SyncProtocol{node: node}
	sp.node.SetStreamHandler(syncProtocol, sp.handleSyncStream)
	return sp
}

//
// handle connection between peers
//
func (sp *SyncProtocol) handleSyncStream(s net.Stream) {

	log.Println("\t\tGot a new stream!")

	// wrap stream for protobuf exchange
	ws := WrapStream(s)

	// go sp.ReadData(ws)
	err := sp.createInboundReader(ws)
	if err != nil {
		log.Println("unable to attach reader to inbound stream: ", err)
	}

	// go sp.WriteData(ws)
	err = sp.createOutboundWriter(ws)
	if err != nil {
		log.Println("unable to attach outbound writer to nss feed: ", err)
	}

	// stream 's' will stay open until you close it (or the other side closes it).
}

//
// for the inbound stream listens for messages
// and adds to the local feed
//
func (sp *SyncProtocol) createInboundReader(ws *WrappedStream) error {

	// create stan connection with random client id
	sc, err := NSSConnection(nuid.Next())
	if err != nil {
		return err
	}

	// launch the read loop, append received blocks to the local feed
	go func() {
		defer sc.Close()
		defer ws.stream.Close()
		log.Println("reader open for: ", string(ws.remotePeerID()))
		i := 0
		for {
			b, err := receiveBlock(ws)
			if err != nil {
				log.Println("read-block error, closing reader: ", err)
				break
			}

			// log.Println("...got a message from: ", string(ws.remotePeerID()))

			// verify - check content against signature
			if !b.Verify() {
				log.Printf("\n\nrecieved block failed verification %v\n\n", b)
				continue
			}
			// log.Println("...received block is verified")

			// validate by attempting to add to blockchain
			bc := GetBlockchain(b.Data.Context, b.Author)
			validBlock, err := bc.AddBlock(b)
			if err != nil {
				log.Println("\t\t=== received invalid block ===")
				b.Print()
				continue
			}
			// log.Println("\t...received block is valid")

			// if all ok publish to the feed
			mutex.Lock()
			filterfeed_records++
			mutex.Unlock()
			err = sc.Publish("feed", validBlock.Serialize())
			if err != nil {
				log.Println("unable to publish message to nss: ", err)
				break
			}
			// log.Println("...inbound message committed to nss:feed")
			i++
			log.Printf("messages received from:\t%s: %d\n", ws.remotePeerID(), i)
			// b.Print()
		}
		// on any errors return & close nats and stream connections
		return
	}()

	return nil
}

//
// attaches to the main feed & sends messages to the
// connected peer
//
func (sp *SyncProtocol) createOutboundWriter(ws *WrappedStream) error {

	// create stan connection with random client id
	pid := string(ws.remotePeerID())
	sc, err := NSSConnection(pid)
	if err != nil {
		return err
	}

	go func() {
		errc := make(chan error)
		defer close(errc)
		defer sc.Close()

		// main message sending routine
		sub, err := sc.Subscribe("feed", func(m *stan.Msg) {

			sendMessage := true

			// get the block from the feed
			blk := DeserializeBlock(m.Data)

			// log.Printf("\n\n\tauthor: %s\n\tremote-peer: %s", blk.Author, ws.remotePeerID())

			// don't send if created by the remote peer
			if blk.Author == ws.remotePeerID() {
				sendMessage = false
				// log.Println("author is remote peer sendmessage:", sendMessage)
			}

			// log.Printf("\n\n\tsender: %s\n\tremote-peer: %s", blk.Sender, ws.remotePeerID())
			// don't send if it came from the reomote peer
			if blk.Sender == ws.remotePeerID() {
				sendMessage = false
				// log.Println("sender is remote peer sendmessage:", sendMessage)
			}

			// don't send if the remote peer has seen it before
			if sendMessage {
				blk.Receiver = ws.remotePeerID()
				if sp.node.msgCMS.Estimate(blk.CmsKey()) > 0 {
					sendMessage = false
					// log.Println("message:", blk.BlockId, " already sent to:", ws.remotePeerID(), " sendmaessage:", sendMessage)
				}
			}

			if sendMessage {
				// update the sender
				blk.Sender = ws.localPeerID()
				sendErr := sendBlock(blk, ws)
				if sendErr != nil {
					log.Println("cannot send message to peer: ", sendErr)
					errc <- sendErr
				}
				// register that we've sent this message to this peer
				sp.node.msgCMS.Update(blk.CmsKey(), 1)
			}

		}, stan.DeliverAllAvailable())
		if err != nil {
			log.Println("error creating feed subscription: ", err)
			return
		}

		// wait for errors on writing, such as stream closing
		<-errc
		sub.Close()
	}()

	return nil
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
// sendBlock encodes and writes a data-message to the outbound stream
//
func sendBlock(block *Block, ws *WrappedStream) error {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered error in send block", r)
		}
	}()

	err := ws.enc.Encode(block)
	if err != nil {
		log.Printf("block encoding error:\n\n%v\n\n", block)
		return err
	}
	// Because output is buffered with bufio, we need to flush!
	ws.w.Flush()
	return err
}
