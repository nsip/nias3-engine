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

	// initiate n3 shutdown handler
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		nss.Stop()
		log.Println("n3 shutdown complete")
		os.Exit(1)
	}()

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

	log.Println("starting p2p node")

	// Make a node that listens on the given multiaddress
	node := n3.NewNode(*listenF, *secio, *seed)

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
			nss.Stop()
			log.Fatal("cannot connect to nss for test mesasges: ", err)
		}
		defer sc.Close()

		for x := 0; x < 5; x++ {
			for i := 0; i < 1; i++ {

				log.Println("generating test message...")

				// build a tuple
				t := &n3.SPOTuple{Context: "SIF", Subject: "Subj", Predicate: "Pred", Object: "Obj", Version: uint64(i)}
				// add it to the blockchain
				b := localBlockchain.AddBlock(t)

				// log.Printf("\t...TestGen\n\n%+v\n\n", b)

				blockBytes := b.Serialize()
				// _ = blockBytes
				err := sc.Publish("feed", blockBytes)
				if err != nil {
					log.Println("cannot send new block to feed: ", err)
					break
				}
				log.Println("...sent a test message to nss:feed")
			}
			time.Sleep(time.Second * 5)
		}
		log.Println("all test messages sent")
	}()

	select {}
}
