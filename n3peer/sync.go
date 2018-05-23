// sync.go

package n3peer

import (
	"context"
	"fmt"
	"log"

	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"

	"github.com/nsip/nias3-engine/messages"
	// pipes "github.com/nsip/nias3-engine/pipelines"
	// f2p "github.com/nsip/nias3-engine/pipelines/feed-to-peer"
	"github.com/pkg/errors"
)

//
// sync protocol provides method handlers for
// syncing data between peers on the network
//

// pattern: /protocol-name/request-or-response-message/version
const syncRequest = "/sync/syncreq/0.0.1"
const syncResponse = "/sync/syncresp/0.0.1"

// SyncProtocol type
type SyncProtocol struct {
	node *Node // local host
}

func NewSyncProtocol(node *Node) *SyncProtocol {
	sp := &SyncProtocol{node: node}
	node.SetStreamHandler(syncRequest, sp.onSyncRequest)
	node.SetStreamHandler(syncResponse, sp.onSyncResponse)
	return sp
}

//
// remote peer sync requests handler
// receives sync request and (if valid/trusted)
// puts the request on the inbound q for processing by
// the relevant main app pipeline
//
func (sp *SyncProtocol) onSyncRequest(s inet.Stream) {

	log.Println("onSyncRequest:\tincoming sync req. stream received...")

	// authenticate
	if !isTrustedPeer(s) {
		return
	}

	// wrap the stream to deal with protobuf exchange
	ws := sp.node.WrapStream(s)
	defer ws.stream.Close()

	// get request data
	sreq, err := receiveSyncRequest(ws)
	if err != nil {
		log.Println("unable to unmarshal sync request: ", err)
		return
	}

	log.Printf("onSyncRequest:\t%s: Received sync request from %s. Message: %s", s.Conn().LocalPeer(), s.Conn().RemotePeer(), sreq)

	log.Println("onSyncRequest:\tcontent digest:")
	for k, v := range sreq.LedgerDigest {
		log.Printf("\t%s : %d\n", k, v)
	}

	// write request into app pipelines
	sp.node.inboundSyncRequestChan <- sreq

}

//
// handles reading a data sync request from the provided stream
//
func receiveSyncRequest(ws *WrappedStream) (*messages.SyncRequest, error) {
	var msg messages.SyncRequest
	err := ws.dec.Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

//
// will be invoked by another peer sending
// data in response to a SyncRequest
//
func (sp *SyncProtocol) onSyncResponse(s inet.Stream) {

	log.Println("onSyncResponse:\tstream received...")
	ws := sp.node.WrapStream(s)
	sp.handleSyncStream(ws)
}

// handleSyncStream is a for loop which receives and then passes
// signed messages on into the main app pipelines.
func (sp *SyncProtocol) handleSyncStream(ws *WrappedStream) {

	// read from the peer stream until an error (ie stream closes EOF)
	for {
		msg, err := receiveSyncMessage(ws)
		if err != nil {
			break
		}
		log.Println("handleSyncStream:\treceived message...")
		// send message to main app replication pipeline
		sp.node.inboundMsgChan <- msg
	}
}

//
// receiveSyncMessage reads and decodes a data message from the stream
//
func receiveSyncMessage(ws *WrappedStream) (*messages.SignedN3Message, error) {
	var msg messages.SignedN3Message
	err := ws.dec.Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

//
// utility method that can be called from app pipelines to send a
// sync request to a remote peer
//
func (sp *SyncProtocol) SendSyncRequest(srq *messages.SyncRequest) error {

	log.Println("SendSyncRequest:\topening stream to remote peer: ", srq.TargetId)

	// make a new stream to remote host
	ws, err := sp.GetWrappedSyncRequestStream(srq.TargetId)
	if err != nil {
		return errors.Wrap(err, "cannot create stream to remote peer:")
	}
	log.Println("SendSyncRequest:\tremote peer stream established")
	log.Println("SendSyncRequest:\tsending sync request...")
	return sendSyncRequest(srq, ws)
}

//
// sends a data sync reuest message
//
func sendSyncRequest(srq *messages.SyncRequest, ws *WrappedStream) error {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered in sendSyncRequest", r)
		}
	}()

	err := ws.enc.Encode(srq)
	// Because output is buffered with bufio, we need to flush!
	ws.w.Flush()
	return err
}

//
// For a given remote peer ID returns a stream for sync request exchange
//
func (sp *SyncProtocol) GetWrappedSyncRequestStream(target string) (*WrappedStream, error) {

	peerid, err := peer.IDB58Decode(target)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode target peer ID:")
	}

	log.Println("GetWrappedSyncRequest:\topening stream to remote peer: ", peerid)

	// make a new stream to remote host
	s, err := sp.node.NewStream(context.Background(), peerid, syncRequest)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create stream to remote peer:")
	}
	log.Println("GetWrappedSyncRequest:\tremote peer stream established")

	ws := sp.node.WrapStream(s)

	return ws, nil
}

//
// For a given remote peer ID returns a stream for sync data-message exchange
//
func (sp *SyncProtocol) GetWrappedSyncResponseStream(target string) (*WrappedStream, error) {

	peerid, err := peer.IDB58Decode(target)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode target peer ID:")
	}

	log.Println("GetWrappedSyncResponse:\topening stream to remote peer: ", peerid)

	// make a new stream to remote host
	s, err := sp.node.NewStream(context.Background(), peerid, syncResponse)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create stream to remote peer:")
	}
	log.Println("GetWrappedSyncResponse:\tremote peer stream established")

	ws := sp.node.WrapStream(s)

	return ws, nil

}

//
// SendSyncResponseMessage encodes and writes a data-message to the outbound stream
//
func (sp *SyncProtocol) SendSyncResponseMessage(msg *messages.SignedN3Message, ws *WrappedStream) error {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered in send response", r)
		}
	}()

	err := ws.enc.Encode(msg)
	// Because output is buffered with bufio, we need to flush!
	ws.w.Flush()
	return err
}

//
// check whether the request comes from someone we trust
// TODO: Abstract to trust service.
//
func isTrustedPeer(s inet.Stream) bool {

	log.Println("trust check: ", s.Conn().RemotePeer().String())

	// TODO - Obviously change this!!!
	return true
}
