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

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	context := flag.String("c", "SIF", "blockchain context")
	inet := flag.Bool("inet", false, "turn on node p2p access to external network")
	flag.Parse()

	// start the streaming server
	nss := n3.NewNSS()
	err := nss.Start()
	if err != nil {
		nss.Stop()
		log.Fatal(err)
	}
	defer nss.Stop()

	// create stan connection for writing to feed
	sc, err := n3.NSSConnection("n3main")
	if err != nil {
		log.Println("cannot connect to nss: ", err)
		return
	}
	defer sc.Close()

	// create the message-dedupe count min sketch
	msgCMS, err := n3.NewN3CMS("./msgs.cms")
	if err != nil {
		msgCMS.Close()
		log.Fatal(err)
	}

	// create/open the default blockchain
	localBlockchain = n3.NewBlockchain(*context)

	// if truly new, commit the genesis block to the feed
	bi := localBlockchain.Iterator()
	b := bi.Next()
	if b.Data.Subject == "Genesis" {
		err := sc.Publish("feed", b.Serialize())
		if err != nil {
			log.Println("cannot commit gnensis block to feed: ", err)
		} else {
			log.Println("genesis block committed to feed")
		}
	}

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	log.Println("starting p2p node")

	// Make a node that listens on the given port for multiaddress:
	// will listen on all local network interfaces
	// if --inet then will also connect to external interface
	node := n3.NewNode(*listenF, *inet, msgCMS)

	// connect to remote peer if supplied
	if *target != "" {
		node.ConnectToPeer(*target)
	}

	// generate some test messages
	go func() {

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

				blockBytes := b.Serialize()
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

	// initiate n3 shutdown handler
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		nss.Stop()
		msgCMS.Close()
		localBlockchain.Close()
		log.Println("n3 shutdown complete")
		os.Exit(1)
	}()

	// wait for shutdown
	select {}
}
