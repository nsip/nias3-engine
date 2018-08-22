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
	var infl *n3.InfluxModel
	var mongo *n3.MongoModel

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
		n3.CloseLocalBlockchains()
		if nss != nil {
			nss.Stop()
		}
		if msgCMS != nil {
			msgCMS.Close()
		}
		if hexa != nil {
			hexa.Close()
		}
		if infl != nil {
			infl.Close()
		}
		if mongo != nil {
			mongo.Close()
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
			log.Println("cannot commit genesis block to feed: ", err)
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

	/*
		Topology of streams and stores:
		(API, peers) -> Blockchain.AddNewBlock() -> "stream:feed"
		All incoming records, from other peers or from local API, go into "feed" via sigchain

		"stream:feed" -> Hexastore.db + N3CMS.cms + FilterFeed() -> "stream:filteredfeed"
		By looking up the in-memory CMS, incoming records are filtered into the records that
		will actually be saved. If any tuple has an empty object, a delete command is issued
		for that tuple (delete the currently stored SPO tuple keys).  These are queued as
		key/value pairs into a write-ahead log, to deal with the slowness of random writes on bolt.

		"stream:filteredfeed" -> Hexastore.ConnectToFeed -> Hexastore.db
		The key/value pairs are fetched from filteredfeed, and stored into the primary Hexastore,
		which is the write model.
		Delete commands are realised as deleting all Hexastore key permutations for the latest
		stored tuple, and tombstoning that tuple. Put commands are realised as putting all
		Hexastore key permutations.
		Filteredfeed is a write-ahead log: entries are fetched every 0.1 sec, because of
		how expensive bolt updates are. Within a batch, tuples are sorted by SP then timestamp:
		only the first command (if a delete) and the last command (if a put) are realised.
	*/

	// start the hexastore
	log.Println("starting hexastore")
	hexa = n3.NewHexastore()
	err = hexa.FilterFeed()
	if err != nil {
		log.Fatal("cannot connect CMS to feed")
	}
	err = hexa.ConnectToFeed()
	if err != nil {
		log.Fatal("cannot connect hexastore to feed")
	}

	infl = n3.NewInfluxModel()
	err = infl.ConnectToFeed()
	if err != nil {
		log.Fatal("cannot connect influx model to feed")
	}

	mongo = n3.NewMongoModel()
	err = mongo.ConnectToFeed()
	if err != nil {
		log.Fatal("cannot connect mongo model to feed")
	}

	// start the webserver
	log.Println("starting webserver")
	go n3.RunWebserver(*webPort, hexa, infl)

	// wait for shutdown
	select {}
}
