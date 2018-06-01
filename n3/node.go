// node.go

package n3

import (
	"bufio"
	"context"
	"fmt"
	"log"

	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	multicodec "github.com/multiformats/go-multicodec"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"
	"github.com/nsip/nias3-engine/utils"
)

// node client version
const clientVersion = "n3/0.0.1"

// Node type - a p2p host implementing one or more p2p protocols
type Node struct {
	host.Host // lib-p2p host
	*SyncProtocol
	msgCMS *N3CMS
}

// streamWrap wraps a libp2p stream. We encode/decode whenever we
// write/read from a stream, so we can just carry the encoders
// and bufios with us
type WrappedStream struct {
	stream net.Stream
	enc    multicodec.Encoder
	dec    multicodec.Decoder
	w      *bufio.Writer
	r      *bufio.Reader
}

//
// helper method returns the string id of the
// connecting peer
//
func (ws *WrappedStream) remotePeerID() string {
	return ws.stream.Conn().RemotePeer().Pretty()
}

func (ws *WrappedStream) localPeerID() string {
	return ws.stream.Conn().LocalPeer().Pretty()
}

// wrapStream takes a stream and complements it with r/w bufios and
// decoder/encoder. In order to write raw data to the stream we can use
// wrap.w.Write(). To encode something into it we can wrap.enc.Encode().
// Finally, we should wrap.w.Flush() to actually send the data. Handling
// incoming data works similarly with wrap.r.Read() for raw-reading and
// wrap.dec.Decode() to decode.
func WrapStream(s net.Stream) *WrappedStream {
	reader := bufio.NewReader(s)
	writer := bufio.NewWriter(s)
	// This is where we pick our specific multicodec. In order to change the
	// codec, we only need to change this place.
	dec := protobufCodec.Multicodec(nil).Decoder(reader)
	enc := protobufCodec.Multicodec(nil).Encoder(writer)

	return &WrappedStream{
		stream: s,
		r:      reader,
		w:      writer,
		enc:    enc,
		dec:    dec,
	}
}

// Create a new node with its implemented protocols
// inet creates peer listening on public network interface
// otherwise will default to localhost and local LAN
// interfaces only
//
func NewNode(lport int, inet bool, msgcms *N3CMS) *Node {

	var port int
	if lport == 0 {
		port = utils.GetAvailPort()
	} else {
		port = lport
	}

	bootstrapPeers := IPFS_PEERS
	host, err := makeRoutedHost(port, bootstrapPeers, inet)
	if err != nil {
		log.Fatal(err)
	}

	node := &Node{Host: host}
	node.SyncProtocol = NewSyncProtocol(node)
	node.msgCMS = msgcms
	return node
}

//
// given a remote peer address, attempts to connect and
// instatiates protocol handlers
//
func (n *Node) ConnectToPeer(peerAddress string) error {

	// The following code extracts target's peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(peerAddress)
	if err != nil {
		return err
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return err
	}

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		return err
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr so we add it to the peerstore
	// so LibP2P knows how to contact it
	n.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

	err = n.initiateSyncProtocol(peerid)

	return err

}

func (n *Node) initiateSyncProtocol(peerid peer.ID) error {

	log.Println("...opening stream to remote peer")
	// initiate a stream to the remote host
	s, err := n.NewStream(context.Background(), peerid, syncProtocol)
	if err != nil {
		return err
	}
	log.Println("...stream established")

	go n.handleSyncStream(s)

	return nil

}
