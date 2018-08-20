// influx.go

package n3

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/nats-io/go-nats-streaming"
	inf "github.com/nsip/n3-transport/n3influx"
	"github.com/nsip/n3-transport/pb"
	"github.com/pkg/errors"
	"gopkg.in/fatih/set.v0"
)

func init() {
}

type InfluxModel struct {
	Pub *inf.Publisher
}

func NewInfluxModel() *InfluxModel {
	n3ic, err := inf.NewPublisher()
	if err != nil {
		log.Fatal(errors.Wrap(err, "cannot open influx read model."))
	}
	return &InfluxModel{
		Pub: n3ic,
	}
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

func filter_batch(commands dbCommandSlice) dbCommandSlice {
	sort.Sort(commands)
	out := make([]*DbCommand, 0)
	tmp := make([]*DbCommand, 0)
	issuedDelete := false
	if debug {
		for _, c := range commands {
			log.Printf("INFLUX INPUT: %s %s %s %s %d\n", c.Verb, c.Data.Subject, c.Data.Predicate, c.Data.Object, c.Sequence)
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
				out = append(out, commands[i])
			}
		}
	}
	return dbCommandSlice(out)
}

// Save a batch of commands to delete or put tuples, and transform them into Influx time series entries
// Tombstones are ignored.
// * Filter out redundant commands
// * Process the commands for an object in order: delete entries, put entries
func (hx *InfluxModel) update_batch(commands dbCommandSlice) error {
	if debug {
		log.Printf("Updating %d entries\n", len(commands))
	}
	if len(commands) == 0 {
		return nil
	}
	out1 := filter_batch(commands)
	for _, cmd := range out1 {
		if debug {
			log.Printf("INFLUX OUT: %s %s %s %s\n", cmd.Verb, cmd.Data.Subject, cmd.Data.Predicate, cmd.Data.Object)
		}
		tuple := &pb.SPOTuple{Subject: cmd.Data.Subject, Object: cmd.Data.Object, Predicate: cmd.Data.Predicate,
			PredicateFlat: cmd.Data.PredicateFlat, Context: cmd.Data.Context, Version: int64(cmd.Data.Version)}
		switch cmd.Verb {
		case "delete":
			hx.Pub.DeleteTuple(tuple)
		case "put":
			hx.Pub.StoreTuple(tuple)
		}
	}
	return nil
}

// Reads tuples to be stored or deleted onto read model, and stores or deletes them.
func (rm *InfluxModel) ConnectToFeed() error {
	// create stan connection for writing to feed
	sc, err := NSSConnection("n3readmodel")
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
			//log.Printf("%s %s %s %s\n", cmd.Verb, cmd.Data.Subject, cmd.Data.Predicate, cmd.Data.Object)
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

// get latest non-tombstoned tuples matching subject and context
func (hx *InfluxModel) GetTuples(subject string, context string) ([]*SPOTuple, error) {
	query := fmt.Sprintf("SELECT subject, predicate, object, tombstone, version from %s WHERE subject = '%s' GROUP BY predicate", context, subject)
	return hx.getInfluxTuples(query, context)
}

// get all xAPI tuples relating to a student identified by SIF Refid
// does join on email addresses, which are what xAPI uses as ID (in test data)
func (hx *InfluxModel) GetXapiTuplesBySIFRefid(refid string) ([]*SPOTuple, error) {
	query := fmt.Sprintf("SELECT subject, predicate, object, tombstone, version from SIF WHERE subject = '%s' AND (predicateflat = 'StaffPersonal.PersonInfo.EmailList.Email' OR predicateflat = 'StudentPersonal.PersonInfo.EmailList.Email')", refid)
	tuples, err := hx.getInfluxTuples(query, "SIF")
	if err != nil || len(tuples) == 0 {
		return make([]*SPOTuple, 0), err
	}
	emails := getInfluxValues(tuples, "object")
	querytail := ""
	for i, e := range emails {
		if i > 0 {
			querytail += " OR "
		}
		querytail += "(predicate = 'actor.mbox' AND object = 'mailto:" + e + "')"
	}
	query = "SELECT subject, predicate, object, tombstone, version from xAPI WHERE (" + querytail + ")"
	tuples, err = hx.getInfluxTuples(query, "xAPI")
	if err != nil || len(tuples) == 0 {
		return make([]*SPOTuple, 0), err
	}
	ids := getInfluxValues(tuples, "subject")
	querytail = ""
	for i, e := range ids {
		if i > 0 {
			querytail += " OR "
		}
		querytail += "subject = '" + e + "'"
	}
	query = "SELECT subject, predicate, object, tombstone, version from xAPI WHERE (" + querytail + ")"
	return hx.getInfluxTuples(query, "xAPI")
}

// get distinct values of subject/object/predicate as a string slice from a slice of tuples
func getInfluxValues(tuples []*SPOTuple, value string) []string {
	ret := set.New()
	for _, t := range tuples {
		switch value {
		case "subject":
			ret.Add(t.Subject)
		case "object":
			ret.Add(t.Object)
		case "predicate":
			ret.Add(t.Predicate)
		}
	}
	return set.StringSlice(ret)
}

// get tuples out of Influx query
// Presupposes query has been for subject, predicate, object, tombstone, version in that order
// Appends "ORDER BY time DESC" to all queries
func (hx *InfluxModel) getInfluxTuples(query string, context string) ([]*SPOTuple, error) {
	objs := make([]*SPOTuple, 0)
	response, err := hx.Pub.Query(influx.Query{Command: query + " ORDER BY time DESC", Database: "tuples"})
	if err != nil {
		log.Println(err)
		return objs, err
	}
	if response.Error() != nil {
		log.Println(response.Error())
		return objs, response.Error()
	}
	if len(response.Results) > 0 {
		for i := range response.Results {
			if len(response.Results[i].Series) > 0 {
				for _, s := range response.Results[i].Series {
					//log.Printf("%#v\n", s)
					if len(s.Values) == 0 {
						continue
					}
					for _, v := range s.Values {
						// gotcha with influx queries: the zeroth result column is always the timestamp
						if v[4].(string) == "true" {
							// tombstoned value
							continue
						}
						if v[3] == nil {
							// deleted object
							continue
						}
						version, err := v[5].(json.Number).Int64()
						if err != nil {
							version = 1
						}
						objs = append(objs, &SPOTuple{
							Subject:   v[1].(string),
							Predicate: v[2].(string),
							Object:    v[3].(string),
							Context:   context,
							Version:   uint64(version),
						})
					}
				}
			}
		}
	}
	return objs, nil
}

//
// clean shutdown of underlying data store
//
func (rm *InfluxModel) Close() {
	//
}
