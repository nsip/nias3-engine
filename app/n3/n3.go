package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	n3 "github.com/nsip/nias3-engine/n3"
)

var localBlockchain *n3.Blockchain

func main() {

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	context := flag.String("c", "SIF", "blockchain context")
	inet := flag.Bool("inet", false, "turn on node p2p access to external network")
	webPort := flag.Int("webport", 1340, "port to run web handler on")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to `file`")
	flag.Parse()

	var nss *n3.NSS
	var msgCMS *n3.N3CMS
	var localBlockchain *n3.Blockchain
	var hexa *n3.Hexastore

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}

	// initiate n3 shutdown handler
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP,
		syscall.SIGINT, syscall.SIGILL, syscall.SIGABRT, syscall.SIGEMT, syscall.SIGSYS)

	go func() {
		<-c
		if nss != nil {
			nss.Stop()
		}
		if msgCMS != nil {
			msgCMS.Close()
		}
		if localBlockchain != nil {
			localBlockchain.Close()
		}
		if hexa != nil {
			hexa.Close()
		}
		if *cpuprofile != "" {
			pprof.StopCPUProfile()
		}
		log.Println("n3 shutdown complete")
		os.Exit(1)
	}()

	// start the streaming server
	nss = n3.NewNSS()
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

	log.Println("starting Count-Minsketch")
	// create the message-dedupe count min sketch
	msgCMS, err = n3.NewN3CMS("./msgs.cms")
	if err != nil {
		msgCMS.Close()
		log.Fatal(err)
	}

	log.Println("starting Sigchain")
	// create/open the default blockchain
	// TODO allow multiple contexts
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

	// start the hexastore
	log.Println("starting hexastore")
	hexa = n3.NewHexastore()
	err = hexa.ConnectToFeed()
	if err != nil {
		log.Fatal("cannot connect hexastore to feed")
	}

	// start the webserver
	log.Println("starting webserver")
	go n3.RunWebserver(*webPort, hexa)

	// wait for shutdown
	select {}
}
