// n3.go

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/nsip/nias3-engine/messages"
	"github.com/nsip/nias3-engine/n3peer"
	pipes "github.com/nsip/nias3-engine/pipelines"
	f2h "github.com/nsip/nias3-engine/pipelines/feed-to-hex"
	i2f "github.com/nsip/nias3-engine/pipelines/ingest-to-feed"
	p2r "github.com/nsip/nias3-engine/pipelines/peer-to-replicate"
	r2f "github.com/nsip/nias3-engine/pipelines/replicate-to-feed"
	sr "github.com/nsip/nias3-engine/pipelines/sync-request"
	t2i "github.com/nsip/nias3-engine/pipelines/tuple-to-ingest"
	"github.com/nsip/nias3-engine/storage/leveldb"
	"github.com/nsip/nias3-engine/stream/n3stan"
)

func main() {

	// Parse options from the command line
	target := flag.String("d", "", "target peer to dial")
	global := flag.Bool("global", true, "use global ipfs peers for bootstrapping")
	name := flag.String("name", "", "id for testing (appears in test data)")
	testMessages := flag.Int("test-messages", 0, "no. test mesasges to generate")
	flag.Parse()

	// start the streaming server
	nss := n3stan.NewNSS()
	err := nss.Start()
	if err != nil {
		nss.Stop()
		log.Fatal(err)
	}
	defer nss.Stop()

	// give the nss time to come up
	time.Sleep(time.Second * 5)

	// initiate n3 shutdown handler
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		nss.Stop()
		log.Println("n3 shutdown complete")
		os.Exit(1)
	}()

	log.Println("==== Creating Main Processor Pipelines ====")

	// -- tuple to ingest pipeline
	dtChan, errcList, err := t2i.TupleToIngestPipeline(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// monitor the pipeline
	go pipes.MonitorPipeline("tuple-to-ingest", errcList...)
	log.Println("tuple-to-ingest created...")

	// -- ingest to feed pipeline
	errcList, err = i2f.IngestToFeedPipeline(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	go pipes.MonitorPipeline("ingest-to-feed", errcList...)
	log.Println("ingest-to-feed created...")

	// -- feed to hex
	errcList, err = f2h.FeedToHexPipeline(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	go pipes.MonitorPipeline("feed-to-hex", errcList...)
	log.Println("feed-to-hex created...")

	// --replicate to feed
	errcList, err = r2f.ReplicateToFeedPipeline(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// monitor the pipeline
	go pipes.MonitorPipeline("replicate-to-feed", errcList...)
	log.Println("replicate-to-feed created...")

	// --peer to replicate
	inboundMsgChan, errcList, err := p2r.PeerToReplicatePipeline(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	go pipes.MonitorPipeline("peer=to-replicate", errcList...)
	log.Println("peer-to-replicate created...")

	//
	// create an inbound entry point for sync-requests
	//
	inboundSyncRequestWriter := n3stan.NewSyncRequestWriter()
	if err != nil {
		log.Fatal(err)
	}
	err = inboundSyncRequestWriter.Subscribe("sr.inbound", "n3app")
	if err != nil {
		log.Fatal(err)
	}
	inboundSyncRequestChan := inboundSyncRequestWriter.MsgC()

	//
	// create the p2p node
	//
	log.Println("=== Creating P2P Node & Pipelines ===")
	p2pnode := n3peer.NewNode(*global, inboundMsgChan, inboundSyncRequestChan)
	// log.Printf("\n\n%#v\n\n", p2pnode)

	//
	// once peer is active
	// create the pipelines that use it
	//

	//
	// handle outbound sync requests
	//
	errcList, err = sr.OutboundSyncRequestPipeline(context.Background(), p2pnode)
	if err != nil {
		log.Fatal(err)
	}
	go pipes.MonitorPipeline("outbound-sync-request", errcList...)

	//
	// handle inbound sync requests
	//
	errcList, err = sr.InboundSyncRequestPipeline(context.Background(), p2pnode)
	if err != nil {
		log.Fatal(err)
	}
	go pipes.MonitorPipeline("inbound-sync-request", errcList...)

	// send in some ingest messages if required
	// constantly growing the local dataset at intervals
	// five iterations to give some variability of load
	// over time
	// go func() {
	if *testMessages > 0 {
		log.Println("=== Initiating local test messages ===")
		for i := 0; i < 2; i++ {
			sendTestIngestMessages(*name, dtChan, *testMessages)
			time.Sleep(time.Second * 1)
			log.Println("\ttest messages generated iteration: ", i+1)
		}
	}
	// }()

	//
	// if a target peer was passed then
	// start the synchronisation p2p exchange
	// by writing an outbound sync request
	//
	if *target != "" {
		err := initiateSync(p2pnode, *target)
		if err != nil {
			log.Println("...cannot initiate sync: ", err)
		}
	}

	select {}

}

//
// constructs a sync request to start the
// syncing conversation with the remote peer (target)
//
func initiateSync(node *n3peer.Node, target string) error {

	ls, err := leveldb.NewLedgerService()
	if err != nil {
		return err
	}

	digest, err := ls.CreateSyncDigest("TEST")
	if err != nil {
		return err
	}

	// The following code extracts target's peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(target)
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

	// note: use this to store in peerstore!!!
	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr so we add it to the peerstore
	// so LibP2P knows how to contact it
	node.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

	outboundSyncRequestWriter := n3stan.NewSyncRequestWriter()
	err = outboundSyncRequestWriter.Subscribe("sr.outbound", "initiateSync")
	if err != nil {
		return err
	}

	syncRequest := &messages.SyncRequest{
		Context:  "TEST",
		SourceId: node.ID().Pretty(),
		// TargetId:     target,
		TargetId:     peerid.Pretty(),
		LedgerDigest: digest,
	}

	log.Println("initiate sync")
	log.Printf("\nrequest:\n\n%s\n\n", syncRequest)

	outboundSyncRequestWriter.MsgC() <- syncRequest

	return nil

}

func sendTestIngestMessages(name string, dtChan chan<- *messages.DataTuple, numMsgs int) {

	for i := 0; i < numMsgs; i++ {

		dt := &messages.DataTuple{
			Subject:   "ingest-AAABB",
			Context:   "TEST",
			Predicate: fmt.Sprintf("attribute-%d-%s", i, name),
			Object:    fmt.Sprintf("value-%d", i)}

		select {
		case dtChan <- dt:
			// log.Println("message sent from pipes endppoint")
		}
	}
	log.Println("all ingest messages sent.")

}
