// mongo.go

package n3

import (
	//"fmt"
	"log"
	"regexp"
	"sort"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

const DBNAME = "readview"

var session *mgo.Session

func init() {
	if session == nil {
		var dbErr error
		session, dbErr = mgo.Dial("localhost")
		if dbErr != nil {
			log.Fatal(errors.Wrap(dbErr, "cannot open mongodb read model."))
		}
	}
}

type MongoModel struct {
	Db      string
	Session *mgo.Session
}

func NewMongoModel() *MongoModel {
	return &MongoModel{
		Db:      DBNAME,
		Session: session,
	}
}

func (hx *MongoModel) DB() *mgo.Database {
	return hx.Session.DB(hx.Db)
}

// Tuple update Commands batch is filtered here as with Hexastore.update_batch(), but without
// generating Hexastore permutations of keys or tombstones:
// * Sort the commands by SP, then timestamp
// * For all commands involving the same SP:
//   * If the first command is a delete, pass it through (do not check what has been saved to disk)
//   * If the last command is a put:
//     * Check whether this is the same tuple as the first deletion request for the batch
//       * If yes, ignore both requests, they do not change the state
//       * Else, pass it through
//   * Ignore all commands in between: they have been overruled within the batch

func (hx *MongoModel) filter_batch(commands dbCommandSlice) dbCommandSlice {
	sort.Sort(commands)
	out := make([]*DbCommand, 0)
	tmp := make([]*DbCommand, 0)
	issuedDelete := false
	if debug {
		for _, c := range commands {
			log.Printf("INPUT: %s %s %s %s %d\n", c.Verb, c.Data.Subject, c.Data.Predicate, c.Data.Object, c.Sequence)
		}
	}
	for i := range commands {
		if commands[i] == nil {
			continue
		}
		// first record in batch of commands with same CSP
		if i == 0 || commands[i-1] == nil ||
			commands[i-1].Data.Context != commands[i].Data.Context ||
			commands[i-1].Data.Context == commands[i].Data.Context &&
				(commands[i-1].Data.Subject != commands[i].Data.Subject ||
					commands[i-1].Data.Subject == commands[i].Data.Subject && commands[i-1].Data.Predicate != commands[i].Data.Predicate) {
			issuedDelete = false
			if commands[i].Verb == "delete" {
				issuedDelete = true
				tmp = make([]*DbCommand, 0)
				tmp = append(tmp, commands[i])
			}
		}
		// last record in batch of commands with same CSP
		if i == len(commands)-1 ||
			commands[i].Data.Context != commands[i+1].Data.Context ||
			commands[i].Data.Context == commands[i+1].Data.Context &&
				(commands[i].Data.Subject != commands[i+1].Data.Subject ||
					commands[i].Data.Subject == commands[i+1].Data.Subject && commands[i].Data.Predicate != commands[i+1].Data.Predicate) {
			if commands[i].Verb == "delete" || issuedDelete {
				// enforce the previous delete
				out = append(out, tmp...)
			}
			if commands[i].Verb == "put" {
				if issuedDelete {
					// ignore this request, it undoes the previous delete
				} else {
					out = append(out, commands[i])
				}
			}
		}
	}
	return dbCommandSlice(out)
}

// Save a batch of commands to delete or put tuples, and transform them into JSON
// documents around the same subject, where applicable. Each document is saved in
// a Mongo collection named for content + (in the case of SIF) root element.
// Tombstones are ignored.
// * Filter out redundant commands
// * Gather up commands by object (same subject and context)
// * Process the commands for an object in order: delete entries, put entries
func (hx *MongoModel) update_batch(commands dbCommandSlice) error {
	if debug {
		log.Printf("Updating %d entries\n", len(commands))
	}
	if len(commands) == 0 {
		return nil
	}
	out1 := hx.filter_batch(commands)
	var tmp []*DbCommand
	for i := range out1 {
		//log.Printf("INPUT: %s %s %s %s %d\n", out1[i].Verb, out1[i].Data.Subject, out1[i].Data.Predicate, out1[i].Data.Object, out1[i].Sequence)
		// new context+subject: gather up all tuples in order to update document with
		if i == 0 || out1[i].Data.Context != out1[i-1].Data.Context ||
			out1[i].Data.Context == out1[i-1].Data.Context && out1[i].Data.Subject != out1[i-1].Data.Subject {
			tmp = make([]*DbCommand, 0)
		}
		tmp = append(tmp, out1[i])
		// last record for the given context+subject: post the tuples
		if i == len(out1)-1 || out1[i].Data.Context != out1[i+1].Data.Context ||
			out1[i].Data.Context == out1[i+1].Data.Context && out1[i].Data.Subject != out1[i+1].Data.Subject {
			if err := hx.update_object(tmp); err != nil {
				return nil
			}
		}
	}
	return nil
}

var rootNodeJsonPath = regexp.MustCompile(`^[^.]+`)

// from NIAS3
var mxj2sjsonPathRe1 = regexp.MustCompile(`\[(\d+)\]`)
var mxj2sjsonPathRe2 = regexp.MustCompile(`\.#text$`)
var mxj2sjsonPathRe3 = regexp.MustCompile(`\.-`)

// from NIAS3
// Convert an MXJ path to a node to an sjson path, using dot notation instead of array, and ".Value" instead of .#text
// For Mongo, convert attribute keys, "-xxx", to "_xxx"
func mxj2sjsonPath(p string) string {
	return mxj2sjsonPathRe1.ReplaceAllString(
		mxj2sjsonPathRe2.ReplaceAllString(
			mxj2sjsonPathRe3.ReplaceAllString(p, "._"),
			".Value"),
		".$1")
}

// update document in mongo with tuples in commands.
// document is identified by subject of the tuples, and is of type given by tuple context + tuple subject,
// which is the same for all commands in the batch
// Document id is tuple subject.
// Document table is context. If context is SIF, document table is context + root element (first node in the predicate),
// which means a table for each SIF object type; SIF is not expected to reuse subjects (RefIDs) across objects.
func (hx *MongoModel) update_object(commands []*DbCommand) error {
	if commands == nil || len(commands) == 0 {
		return nil
	}
	id := commands[0].Data.Subject
	table := commands[0].Data.Context
	if table == "SIF" {
		table = table + "_" + rootNodeJsonPath.FindString(commands[0].Data.Predicate)
	}
	c := hx.DB().C(table)
	var result *bson.M
	doc := []byte(`{"_id": "` + id + `"}`)
	err := c.FindId(id).One(&result)
	if err == nil {
		doc, err = bson.MarshalJSON(result)
		if err != nil {
			log.Println(err)
			return err
		}
		//log.Printf("RETRIEVED: %s\n", string(doc))
		err = c.RemoveId(id)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	for _, cmd := range commands {
		switch cmd.Verb {
		case "delete":
			doc, err = sjson.DeleteBytes(doc, mxj2sjsonPath(cmd.Data.Predicate))
		case "put":
			doc, err = sjson.SetBytes(doc, mxj2sjsonPath(cmd.Data.Predicate), cmd.Data.Object)
		}
	}
	//log.Println(string(doc))
	var m map[string]interface{}
	err = bson.UnmarshalJSON(doc, &m)
	if err != nil {
		log.Println(err)
		return err
	}
	err = c.Insert(m)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Reads tuples to be stored or deleted onto read model, and stores or deletes them.
func (rm *MongoModel) ConnectToFeed() error {
	// create stan connection for writing to feed
	sc, err := NSSConnection("n3mongomodel")
	if err != nil {
		log.Println("cannot connect read model to nss: ", err)
		return err
	}
	log.Println("read model connection to feed ok")

	errc := make(chan error)
	commands := make([]*DbCommand, 0)

	// send batches of received tuples to database
	go func() {
		for {
			if len(commands) > 0 {
				mutex.Lock()
				commands1 := dbCommandSlice(commands)
				commands = make([]*DbCommand, 0)
				mutex.Unlock()
				err = rm.update_batch(commands1)
				if err != nil {
					errc <- errors.Wrap(err, "unable to update hexastore")
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		defer close(errc)
		defer sc.Close()

		// main message handling routine
		sub, err := sc.Subscribe("filteredfeed", func(m *stan.Msg) {
			filtered_records++
			// get the block from the feed
			cmd := DeserializeDbCommand(m.Data)
			//log.Printf("INPUT: %s %s %s %s\n", cmd.Verb, cmd.Data.Subject, cmd.Data.Predicate, cmd.Data.Object)
			mutex.Lock()
			commands = append(commands, cmd)
			mutex.Unlock()
		}, stan.DeliverAllAvailable())
		if err != nil {
			errc <- errors.Wrap(err, "error creating read model feed subscription: ")
		}
		// wait for errors on reading feed or committing tuples
		err = <-errc
		log.Println("error in read model subscription: ", err)
		sub.Close()
		log.Println("read model disconnected from feed")
	}()

	return nil

}

//
// clean shutdown of underlying data store
//
func (rm *MongoModel) Close() {
	rm.Session.Close()
}
