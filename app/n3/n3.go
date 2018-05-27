package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	n3 "github.com/nsip/nias3-engine/n3"
)

var localBlockchain *n3.Blockchain

func main() {

	// create/open the default blockchain
	localBlockchain = n3.NewBlockchain("SIF")

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	secio := flag.Bool("secio", false, "enable secio")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	// Make a node that listens on the given multiaddress
	node := n3.NewNode(*listenF, *secio, *seed)

	// start the web interface

	if *target != "" {

		// The following code extracts target's peer ID from the
		// given multiaddress
		ipfsaddr, err := ma.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}

		pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}

		peerid, err := peer.IDB58Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}

		// Decapsulate the /ipfs/<peerID> part from the target
		// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		// We have a peer ID and a targetAddr so we add it to the peerstore
		// so LibP2P knows how to contact it
		node.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

		log.Println("...opening stream to remote peer")
		// initiate a stream to the remote host
		s, err := node.NewStream(context.Background(), peerid, "/n3/sync/0.0.1")
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("...stream established")

		ws := n3.WrapStream(s)

		go node.ReadData(ws)
		go node.WriteData(ws)

	}

	// generate test messages
	for i := 0; i < 5; i++ {
		// build a tuple
		t := &n3.SPOTuple{Context: "SIF", Subject: "Subj", Predicate: "Pred", Object: "Obj"}
		// add it to the blockchain
		b := localBlockchain.AddBlock(t)
		// send to connected peers
		log.Println("sending message...")
		node.BlockChan <- b
		log.Println("...sent a message")
	}
	log.Println("all messages sent")

	select {}
}
