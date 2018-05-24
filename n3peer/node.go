// node.go

package n3peer

import (
	"bufio"
	"log"

	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	multicodec "github.com/multiformats/go-multicodec"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"

	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/utils"
)

//
// node is the p2p service that manages exchange of n3 data
// between peers on the network
//

// node client version
const clientVersion = "n3-p2p-node/0.0.1"

// Node type - a p2p host implementing one or more p2p protocols
type Node struct {
	host.Host              // lib-p2p host
	inboundMsgChan         chan<- *messages.SignedN3Message
	inboundSyncRequestChan chan<- *messages.SyncRequest
	*SyncProtocol          // sync protocol impl
	// *TrustProtocol // trust protocol impl
}

// streamWrap wraps a libp2p stream. We encode/decode whenever we
// write/read from a stream, so we can just carry the encoders
// and bufios with us
type WrappedStream struct {
	stream inet.Stream
	enc    multicodec.Encoder
	dec    multicodec.Decoder
	w      *bufio.Writer
	r      *bufio.Reader
}

// wrapStream takes a stream and complements it with r/w bufios and
// decoder/encoder. In order to write raw data to the stream we can use
// wrap.w.Write(). To encode something into it we can wrap.enc.Encode().
// Finally, we should wrap.w.Flush() to actually send the data. Handling
// incoming data works similarly with wrap.r.Read() for raw-reading and
// wrap.dec.Decode() to decode.
func (n *Node) WrapStream(s inet.Stream) *WrappedStream {
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
// global: allows node discovery without IP addresses
// inboundChan: connector into the blockchain/hexastore environment
// to pass in any messages received over the p2p protocol
//
func NewNode(global bool, lport int, inboundMsgChan chan<- *messages.SignedN3Message,
	inboundSyncReqChan chan<- *messages.SyncRequest) *Node {

	var port int
	if lport == 0 {
		port = utils.GetAvailPort()
	} else {
		port = lport
	}

	// Make a host that listens on the given multiaddress
	var bootstrapPeers []pstore.PeerInfo
	var globalFlag string
	if global {
		log.Println("using global bootstrap")
		bootstrapPeers = IPFS_PEERS
		globalFlag = " -global"
	} else {
		log.Println("using local bootstrap")
		bootstrapPeers = getLocalPeerInfo()
		globalFlag = ""
	}
	// _ = bootstrapPeers
	// _ = globalFlag
	host, err := makeRoutedHost(port, bootstrapPeers, globalFlag)
	// host, err := makeBasicHost(port)
	if err != nil {
		log.Fatal(err)
	}

	node := &Node{Host: host}
	node.inboundMsgChan = inboundMsgChan
	node.inboundSyncRequestChan = inboundSyncReqChan
	node.SyncProtocol = NewSyncProtocol(node)
	return node
}
