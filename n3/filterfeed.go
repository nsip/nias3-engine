// filterfeed.go

package n3

import (
	"log"

	"github.com/coreos/bbolt"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
)

// for reference only, real message definitions
// are in the protobuf files bloc.proto / block.pb.go

//type DbCommand struct {
//	verb  string // "put, delete"
//	data   SPOTuple
//	sequence uint64
//}

// WARNING
// Speed of STAN: STAN Publish() is not fast, as it waits for an Ack from the
// STAN server before publishing the next message. (Is implemented as a blocking
// call to PublishAsync()
// PublishAsync() is significantly faster, because it does not wait for an Ack;
// but it does not preserve the ordering of messages, as STAN assigns timestamps
// and sequence numbers from the server, not from the client, and PublishAsync()
// does not guarantee the messages will arrive at the server in order
// 80k tuples: sync publishing to feed and filterfeed streams: 62 sec
// 80k tuples: async publishing to feed and filterfeed streams: 17 sec
// Given that peer-to-peer tuple sharing will also be slow, we will put up with this.

// The alternative is to maintain a persistent sequence counter for all tuples
// posted to STAN feed stream (in the blockchain presumably), and to ensure that
// any batch of tuples sent out for storage in a database has no gaps in its
// sequence numbers, which could be a delete/put countering a put/delete.
// (So when sending a batch of tuples for processing in update_batch, don't send the whole
// batch: truncate the batch at the first skip in sequence numbers, and we will wait
// for the missing tuple to be supplied in time for the next iteration of processing).
// We might come back and do this later.

var debug = false

func (c *DbCommand) Serialize() []byte {
	out, err := proto.Marshal(c)
	if err != nil {
		log.Println("tuple-serialize: protobuf encoding error: ", err)
	}
	return out
}

func DeserializeDbCommand(d []byte) *DbCommand {

	command := &DbCommand{}

	err := proto.Unmarshal(d, command)
	if err != nil {
		log.Println("db-command-deserialize: protobuf decoding error: ", err)
	}

	return command

}

// Attaches the hexastore listener / conflict resolver to
// the n3 stan feed, and filters out tuples already seen (via CMS).
// Passes on tuples to be saved onto "filterfeed" stream.
// Indicates deletion of tuples where required as "delete" verb + SPO;
// the O value will be ignored, and interpreted as "delete SP with
// latest stored O value"

func (hx *Hexastore) FilterFeed() error {

	// create stan connection for writing to feed
	sc, err := NSSConnection("n3filter")
	if err != nil {
		log.Println("cannot connect filter to nss: ", err)
		return err
	}
	log.Println("hexa connection to feed ok")

	// create the tuple-dedupe count min sketch
	hexaCMS, err := NewN3CMS("./hexastore.cms")
	if err != nil {
		log.Println("cannot create tuple cms: ", err)
		hexaCMS.Close()
		return err
	}

	records := 0

	go func() {
		errc := make(chan error)
		defer close(errc)
		defer sc.Close()
		defer hexaCMS.Close()

		/*
			ackHandler := func(ackedNuid string, err error) {
				if err != nil {
					log.Printf("Warning: error publishing msg id onto filteredfeed%s: %v\n", ackedNuid, err.Error())
				}
			}
		*/

		// main message handling routine
		sub, err := sc.Subscribe("feed", func(m *stan.Msg) {
			//timestamp := ksuid.New()
			// we are not using timestamps, but sequence numbers; we will multiply seq number by two, to interleave
			// deletes and putsA
			timestamp := m.Sequence * 2
			records++
			if records == 1000 {
				//log.Println("Processed 1000 tuples")
				records = 0
			}
			commitTuple := false
			// get the block from the feed
			blk := DeserializeBlock(m.Data)

			// get the tuple
			t := blk.Data

			// assign data tuple version within this b/c context
			cmsKey := t.CmsKeySP()
			tVer := t.Version
			lastVer := hexaCMS.Estimate(cmsKey)
			if debug {
				log.Printf("%s tver %d lastVer %d %d\n", t.CmsKeySPO(), tVer, lastVer, timestamp)
			}
			var lastEntry *SPOTuple
			lastEntry = nil

			switch {
			case lastVer < tVer:
				commitTuple = true
			case lastVer == tVer:
				var lastEntryBytes []byte
				// get the currently stored tuple. This is a gamble, because the currently
				// stored tuple may not be stored yet to database (we batch them for efficiency)
				err := hx.db.View(func(tx *bolt.Tx) error {
					bkt := NewHexaBucket(tx)
					lastEntryBytes = bkt.Get([]byte(cmsKey))
					return err
				})
				if err != nil {
					errc <- errors.Wrap(err, "Hexa lookup error: ")
				}
				// if there is one do the necessary comparison
				if lastEntryBytes != nil {
					lastEntry = DeserializeTuple(lastEntryBytes)
					// check object legths: arbitrary lexical ordering for colliding tuples with
					// identical version number, as done in gunDB
					if t.Object > lastEntry.Object {
						commitTuple = true
					}
				} else {
					commitTuple = true
				}
			}

			if commitTuple {
				commands := make([]*DbCommand, 0)
				// there is potentially a previous entry to be deleted: we will issue delete command whether it exists or not
				if lastVer > 0 {
					// delete the old entry (ignore the t.Object value: delete the currently saved Object value for t.Subject, t.Predicate
					// set timestamps with offsets: order of execution of the delete and the add are not guaranteed
					commands = append(commands, &DbCommand{Verb: "delete", Sequence: timestamp, Data: t})
				}
				// store the new entry
				commands = append(commands, &DbCommand{Verb: "put", Sequence: timestamp + 1, Data: t})
				if err != nil {
					errc <- errors.Wrap(err, "unable to access hexastore")
				} else {
					for _, command := range commands {
						//log.Printf("%s %s %sn", command.Verb, string(command.Key), string(command.Value))
						err = sc.Publish("filteredfeed", command.Serialize())
						//b, _ := ksuid.FromBytes(command.Ksuid)
						//log.Printf("PUBLISH: %s %s %s\n", command.Verb, string(command.Key), b.String())
						//nuid, err := sc.PublishAsync("filteredfeed", command.Serialize(), ackHandler)
						if err != nil {
							//errc <- errors.Wrap(err, "unable to access message "+nuid+" to nss")
							errc <- errors.Wrap(err, "unable to access message to nss")
							break
						}
					}
				}
				hexaCMS.Update(cmsKey, (tVer - lastVer)) // frequency increment
				// log.Printf("committed tuple: %+v", t)
			} else {
				log.Printf("not committing tuple: %+v; version %d", t, lastVer)
			}
		}, stan.DeliverAllAvailable())
		if err != nil {
			errc <- errors.Wrap(err, "error creating hexastore feed subscription: ")
		}

		// wait for errors on reading feed or committing tuples
		err = <-errc
		log.Println("error in hexa subscription: ", err)
		sub.Close()
		hexaCMS.Close()
		log.Println("hexastore disconnected from feed")
	}()

	return nil

}
