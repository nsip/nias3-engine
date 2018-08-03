// hexastore.go

package n3

import (
	"bytes"
	//"errors"
	"fmt"
	"log"
	//"strconv"
	"sort"
	"sync"
	"time"

	"github.com/coreos/bbolt"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
)

const hexaBucket = "hexa"

var hexboltDB *bolt.DB

func init() {
	if hexboltDB == nil {
		var dbErr error
		hexboltDB, dbErr = bolt.Open("n3hex.db", 0600, &bolt.Options{NoFreelistSync: true})
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

type HexaBucket struct {
	bkt *bolt.Bucket
}

func (bkt *HexaBucket) Bucket() *bolt.Bucket {
	return bkt.bkt
}

func (bkt *HexaBucket) Get(key []byte) []byte {
	return bkt.bkt.Get(key)
}

func (bkt *HexaBucket) Put(key []byte, value []byte) error {
	return bkt.bkt.Put(key, value)
}

func (bkt *HexaBucket) Delete(key []byte) error {
	return bkt.bkt.Delete(key)
}

func (bkt *HexaBucket) Cursor() *bolt.Cursor {
	return bkt.bkt.Cursor()
}

func NewHexaBucket(tx *bolt.Tx) *HexaBucket {
	return &HexaBucket{bkt: tx.Bucket([]byte(hexaBucket))}
}

const CommandLen = 1000

type dbCommandSlice []*DbCommand

func (v dbCommandSlice) Len() int      { return len(v) }
func (v dbCommandSlice) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

func (v dbCommandSlice) Less(i, j int) bool {
	if v[i] == nil {
		return false
	}
	if v[j] == nil {
		return true
	}
	return bytes.Compare(v[i].Key, v[j].Key) == -1 || bytes.Compare(v[i].Ksuid, v[j].Ksuid) == -1
}

func (hx *Hexastore) update_batch(commands []*DbCommand) ([]*DbCommand, time.Time, error) {
	log.Printf("Updating %d entries\n", len(commands))
	if len(commands) == 0 {
		return commands, time.Now(), nil
	}
	err := hx.db.Update(func(tx *bolt.Tx) error {
		bkt := NewHexaBucket(tx)
		commands1 := dbCommandSlice(commands)
		sort.Sort(commands1)
		for _, cmd := range commands1 {
			if cmd == nil {
				continue
			}
			//log.Printf("%#v\n", cmd)
			switch cmd.Verb {
			case "put":
				if err1 := bkt.Put(cmd.Key, cmd.Value); err1 != nil {
					return err1
				}
			case "delete":
				if err1 := bkt.Delete(cmd.Key); err1 != nil {
					return err1
				}
			}
		}
		return nil
	})
	return make([]*DbCommand, 0), time.Now(), err
}

//
// Attaches the hexastore listener / conflict resolver to
// the n3 stan feed, and populates with tuples read from the feed.
// Tuples are stored under all hexastore key permutations.
// Presupposes that there can only be one O value stored under SPO;
// any incoming SPO1 values are treated as conflicting with the original SPO,
// and overwrite that tuple, with the original SPO tombstoned (under a different
// S). Any tuple encoding relating the same S to different O through different P
// needs to differentiate the P instances; in the case of XML encodings, that is
// done by positional encoding of P as a path (e.g. List.Entry.0 vs List.Entry.1).
// Any use of the hexastore for queries presupposing the same P will then need to
// operate on a view of the hexastore, with the variant P1, P2, ... instances
// replaced by a/ unified P.
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

	errc := make(chan error)
	commands := make([]*DbCommand, 0)
	lastUpdate := time.Now()
	var mutex = &sync.Mutex{}
	go func() {
		for {
			if len(commands) > 0 {
				mutex.Lock()
				commands, lastUpdate, err = hx.update_batch(commands)
				mutex.Unlock()
				if err != nil {
					errc <- errors.Wrap(err, "unable to update hexastore")
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	records := 0
	go func() {
		defer close(errc)
		defer sc.Close()
		defer hexaCMS.Close()

		// main message handling routine
		sub, err := sc.Subscribe("filteredfeed", func(m *stan.Msg) {
			records++
			if records == 1000 {
				//log.Println("Filtered Feed: Processed 1000 tuples")
				records = 0
			}
			// get the block from the feed
			cmd := DeserializeDbCommand(m.Data)
			//log.Printf("%s %s\n", cmd.Verb, string(cmd.Key))
			commands = append(commands, cmd)

			//		if len(commands) > CommandLen {
			//			mutex.Lock()
			//			commands, lastUpdate, err = hx.update_batch(commands)
			//			mutex.Unlock()
			//			if err != nil {
			//				errc <- errors.Wrap(err, "unable to update hexastore")
			//			}
			//		}
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

// Store tuple t in Hexastore, under all hexastore permutations of the key of k.
// presupposes hx.db.Update already running
func (bkt *HexaBucket) Store(t *SPOTuple, k *SPOTuple) error {
	var err error
	payload := t.Serialize()
	if err = bkt.Put([]byte(k.CmsKeySPO()), payload); err != nil {
		return err
	}
	if err = bkt.Put([]byte(k.CmsKeySOP()), payload); err != nil {
		return err
	}
	if err = bkt.Put([]byte(k.CmsKeyPSO()), payload); err != nil {
		return err
	}
	if err = bkt.Put([]byte(k.CmsKeyPOS()), payload); err != nil {
		return err
	}
	if err = bkt.Put([]byte(k.CmsKeyOPS()), payload); err != nil {
		return err
	}
	if err = bkt.Put([]byte(k.CmsKeyOSP()), payload); err != nil {
		return err
	}
	if err = bkt.Put([]byte(k.CmsKeySP()), payload); err != nil {
		return err
	}
	return nil
}

// Store tuple t in Hexastore, under only the SPO key of k.
// presupposes hx.db.Update already running
func (bkt *HexaBucket) StoreSingleKey(t *SPOTuple, k *SPOTuple) error {
	var err error
	payload := t.Serialize()
	if err = bkt.Put([]byte(k.CmsKeySPO()), payload); err != nil {
		return err
	}
	return nil
}

// Delete entries in Hexastore, under all hexastore permutations of the key of k.
// presupposes hx.db.Update already running
func (bkt *HexaBucket) Unstore(k *SPOTuple) error {
	var err error
	// log.Printf("Deleting: %s\n", k.CmsKey())
	if err = bkt.Delete([]byte(k.CmsKeySPO())); err != nil {
		return err
	}
	if err = bkt.Delete([]byte(k.CmsKeySOP())); err != nil {
		return err
	}
	if err = bkt.Delete([]byte(k.CmsKeyPSO())); err != nil {
		return err
	}
	if err = bkt.Delete([]byte(k.CmsKeyPOS())); err != nil {
		return err
	}
	if err = bkt.Delete([]byte(k.CmsKeyOPS())); err != nil {
		return err
	}
	if err = bkt.Delete([]byte(k.CmsKeyOSP())); err != nil {
		return err
	}
	if err = bkt.Delete([]byte(k.CmsKeySP())); err != nil {
		return err
	}
	return nil
}

// Get the first entry matching the keyPrefix
func (hx *Hexastore) GetFirstMatching(keyPrefix []byte) ([]byte, error) {
	var k, ret []byte
	err := hx.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(hexaBucket)).Cursor()
		k, ret = c.Seek(keyPrefix)
		if k == nil {
			// log.Printf("%s: no matching prefix\n", string(keyPrefix))
			return errors.New("No matching prefix")
		}
		if !bytes.HasPrefix(k, keyPrefix) {
			// log.Printf("%s: prefix does not match retrieved %s\n", string(keyPrefix), string(k))
			return errors.New("Prefix does not match retrieved entry")
		}
		return nil
	})
	if err != nil {
		log.Println("Iterator error: ", err)
		return nil, err
	}
	// log.Printf("GetFirstMatching: %s\n", string(ret))
	return ret, nil
}

// Check whether key-prefix is on the Hexastore
func (hx *Hexastore) HasKey(keyPrefix string) (bool, error) {
	ret := false
	searchKey := []byte(keyPrefix)
	// log.Printf("search_key: %s\n\n", searchKey)
	err := hx.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(hexaBucket)).Cursor()
		k, _ := c.Seek(searchKey)
		ret = k != nil && bytes.HasPrefix(k, searchKey)
		return nil
	})
	if err != nil {
		log.Println("Iterator error: ", err)
		return ret, err
	}
	return ret, nil
}

// Given a key-prefix, returns the reference ids that
// can be used in a Get operation to retreive the
// desired value

func (hx *Hexastore) GetIdentifiers(keyPrefix string) ([][]byte, error) {
	objIDs := make([][]byte, 0)
	searchKey := []byte(keyPrefix)
	// log.Printf("search_key: %s\n\n", searchKey)
	err := hx.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(hexaBucket)).Cursor()
		for k, v := c.Seek(searchKey); k != nil && bytes.HasPrefix(k, searchKey); k, v = c.Next() {
			objIDs = append(objIDs, v)
			// log.Printf("KEY %s VALUE %s\n", string(k), string(v))
		}
		return nil
	})
	if err != nil {
		log.Println("Iterator error: ", err)
		return nil, err
	}
	return objIDs, nil
}

// given a key-prefix, return all matching tuples
// key prefix is: c:"%s" X:"%s" (Y:"%s" (Z:%s")), where c introduces the context,
// and X Y Z are any of s p o (subject, predicate, object)
func (hx *Hexastore) GetTuples(keyPrefix string) ([]*SPOTuple, error) {
	//log.Println(keyPrefix)
	ids, err := hx.GetIdentifiers(keyPrefix)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	objs := make([]*SPOTuple, 0)
	for _, x := range ids {
		t := DeserializeTuple(x)
		if t.Object != "" {
			objs = append(objs, t)
		}
	}
	return objs, nil
}

// generate a list of all possible keys of a triple
func PermuteTriple(t SPOTuple) [][]byte {
	ret := make([][]byte, 0)
	ret = append(ret, []byte(t.CmsKey()))
	ret = append(ret, []byte(t.CmsKeySOP()))
	ret = append(ret, []byte(t.CmsKeyPSO()))
	ret = append(ret, []byte(t.CmsKeyPOS()))
	ret = append(ret, []byte(t.CmsKeyOPS()))
	ret = append(ret, []byte(t.CmsKeyOSP()))
	return ret
}

// generate a list of all possible keys of a list of triples
func PermuteTripleKeys(list []SPOTuple) [][]byte {
	ret := make([][]byte, 0)
	for _, triple := range list {
		keys := PermuteTriple(triple)
		for _, key := range keys {
			ret = append(ret, key)
		}
	}
	return ret
}

// Generate a tombstoned version of a tuple, flagged it as a delete update:
// see http://thelastpickle.com/blog/2016/07/27/about-deletes-and-tombstones.html
// The tombstoned tuple is kept in the hexastore, but is rendered inaccessible
// by prefixing its subject with a datestamp. (If this is not robust enough,
// will need to add deletion flag to tuples, as Apache Cassandra does.)
func (t *SPOTuple) Tombstone() *SPOTuple {
	ret := SPOTuple{Subject: t.Subject, Object: t.Object, Predicate: t.Predicate, Context: t.Context, Version: t.Version}
	timestamp := "TIMESTAMP_"
	timestampBytes, err := time.Now().MarshalText()
	if err == nil {
		timestamp = fmt.Sprintf("TIMESTAMP_%s_", string(timestampBytes))
	}
	ret.Subject = timestamp + ret.Subject
	return &ret
}

//
// clean shutdown of underlying data store
//
func (hx *Hexastore) Close() {
	hx.db.Close()
}
