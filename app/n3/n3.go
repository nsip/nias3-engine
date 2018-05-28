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

		node.ConnectToPeer(*target)

	}

	// generate test messages
	go func() {
		// create stan connection with test client id
		sc, err := n3.NSSConnection("testWriter")
		if err != nil {
			nss.Stop()
			log.Fatal("cannot connect to nss for test mesasges: ", err)
		}
		defer sc.Close()

		for x := 0; x < 5; x++ {
			for i := 0; i < 5; i++ {
				// build a tuple
				t := &n3.SPOTuple{Context: "SIF", Subject: "Subj", Predicate: "Pred", Object: "Obj"}
				// add it to the blockchain
				b := localBlockchain.AddBlock(t)
				// send to connected peers
				log.Println("sending message...")
				log.Printf("\n\n%s\n\n", b)
				// node.BlockChan <- b
				err := sc.Publish("feed", b.Serialize())
				if err != nil {
					log.Println("cannot send new block to feed: ", err)
					break
				}
				log.Println("...sent a message to feed")
			}
			time.Sleep(time.Second * 10)
		}
		log.Println("all messages sent")
	}()

	select {}
}
