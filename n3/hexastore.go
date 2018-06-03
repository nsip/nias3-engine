// hexastore.go

package n3

import (
	"log"

	"github.com/coreos/bbolt"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
)

const hexaBucket = "hexa"

var hexboltDB *bolt.DB

func init() {
	if hexboltDB == nil {
		var dbErr error
		hexboltDB, dbErr = bolt.Open("n3hex.db", 0600, nil)
		if dbErr != nil {
			log.Fatal(errors.Wrap(dbErr, "cannot open n3 hexstore."))
		}
		// create the hexa bucket in advance
		err := hexboltDB.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(hexaBucket))
			return err
		})
		if err != nil {
			log.Fatal("cannot create hexa bucket in datastore:", err)
		}
	}
}

type Hexastore struct {
	db *bolt.DB
}

func (hx *Hexastore) DB() *bolt.DB {
	return hx.db
}

func NewHexastore() *Hexastore {
	return &Hexastore{db: hexboltDB}
}

//
// attaches the hexastore listener / conflict resolver to
// the n3 stan feed, and populates with tuples read from the feed
//
func (hx *Hexastore) ConnectToFeed() error {

	// create stan connection for writing to feed
	sc, err := NSSConnection("n3hexa")
	if err != nil {
		log.Println("cannot connect hexastore to nss: ", err)
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

	go func() {
		errc := make(chan error)
		defer close(errc)
		defer sc.Close()
		defer hexaCMS.Close()

		// main message hndling routine
		sub, err := sc.Subscribe("feed", func(m *stan.Msg) {

			commitTuple := false

			// get the block from the feed
			blk := DeserializeBlock(m.Data)

			// get the tuple
			t := blk.Data

			// assign data tuple version within this b/c context
			cmsKey := t.CmsKey()
			tVer := t.Version
			lastVer := hexaCMS.Estimate(cmsKey)

			switch {
			case lastVer < tVer:
				commitTuple = true
			case lastVer == tVer:
				var lastEntryBytes []byte
				// get the currently stored tuple
				err := hx.db.View(func(tx *bolt.Tx) error {

					bkt := tx.Bucket([]byte(hexaBucket))
					lastEntryBytes = bkt.Get([]byte(cmsKey))

					return err
				})
				if err != nil {
					errc <- errors.Wrap(err, "Hexa lookup error: ")
				}
				// if there is one do the necessary comparison
				if lastEntryBytes != nil {
					lastEntry := DeserializeTuple(lastEntryBytes)
					// check object legths
					if t.Object > lastEntry.Object {
						commitTuple = true
					}
				}
			}

			if commitTuple {
				// inflate the tuple to hexa entries

				// put the tuple batch in the store
				err := hx.db.Update(func(tx *bolt.Tx) error {
					bkt := tx.Bucket([]byte(hexaBucket))
					err = bkt.Put([]byte(cmsKey), t.Serialize())
					if err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					errc <- errors.Wrap(err, "unable to update hexastore")
				}
				// register the current known version
				hexaCMS.Update(cmsKey, (tVer - lastVer))
				log.Printf("committed tuple: %+v", t)
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

//
// clean shutdown of underlying data store
//
func (hx *Hexastore) Close() {
	hx.db.Close()
}
