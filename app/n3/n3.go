package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	n3 "github.com/nsip/nias3-engine/n3"
)

var localBlockchain *n3.Blockchain

func main() {

	// start the streaming server
	nss := n3.NewNSS()
	err := nss.Start()
	if err != nil {
		nss.Stop()
		log.Fatal(err)
	}
	defer nss.Stop()

	// create the message-dedupe count min sketch
	msgCMS, err := n3.NewN3CMS("./msgs.cms")
	if err != nil {
		msgCMS.Close()
		log.Fatal(err)
	}

	// initiate n3 shutdown handler
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		nss.Stop()
		msgCMS.Close()
		log.Println("n3 shutdown complete")
		os.Exit(1)
	}()

	// create/open the default blockchain
	localBlockchain = n3.NewBlockchain("SIF")
	bi := localBlockchain.Iterator()
	gb := bi.Next()

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	log.Println("starting p2p node")

	// Make a node that listens on the given multiaddress
	node := n3.NewNode(*listenF, msgCMS)

	// connect to remote peer if supplied
	if *target != "" {
		node.ConnectToPeer(*target)
	}

	// time.Sleep(time.Second * 5)

	// generate test messages
	go func() {
		log.Println("creating feed topic")
		// create stan connection with test client id
		sc, err := n3.NSSConnection("testWriter")
		if err != nil {
			log.Println("cannot connect to nss for test mesasges: ", err)
			return
		}
		defer sc.Close()

		// always publish the genesis block
		err = sc.Publish("feed", gb.Serialize())
		if err != nil {
			log.Println("cannot publish genesis block to feed: ", err)
			return
		}

		for x := 0; x < 1; x++ {
			for i := 0; i < 5; i++ {

				log.Println("generating test message...")

				// build a tuple
				t := &n3.SPOTuple{Context: "SIF", Subject: "Subj", Predicate: "Pred", Object: "Obj", Version: uint64(i)}
				// add it to the blockchain
				b, err := localBlockchain.AddNewBlock(t)
				if err != nil {
					log.Println("error adding test data block:", err)
					// break
				}
				log.Println("test message is validated")
				// log.Printf("\t...TestGen\n\n%+v\n\n", b)

				blockBytes := b.Serialize()
				// _ = blockBytes
				err = sc.Publish("feed", blockBytes)
				if err != nil {
					log.Println("cannot send new block to feed: ", err)
					break
				}
				log.Println("...sent a test message to nss:feed")
			}
			time.Sleep(time.Second * 1)
		}
		log.Println("all test messages sent")
	}()

	select {}
}
