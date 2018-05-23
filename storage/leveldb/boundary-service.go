// ledger-service.go

package leveldb

import (
	"fmt"
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type BoundaryService struct {
	client *boundaryClient
}

// Ensure BoundaryService implements n3.BoundaryService interface
var _ n3.BoundaryService = &BoundaryService{}

func NewBoundaryService() (*BoundaryService, error) {

	bs := &BoundaryService{}
	client, err := newBoundaryClient()
	if err != nil {
		return nil, err
	}
	bs.client = client

	return bs, nil

}

//
// finds the next available boundary sequnce number for this
// data tuple and assigns it to the given message
//
func (bs *BoundaryService) AssignBoundary(msg *messages.N3Message) error {

	db := bs.client.db

	last, err := bs.GetLast(msg)
	if err != nil {
		return err
	}

	// var bndry uint64
	bndry := last + 1

	// write the new value for this tuple back to the db
	buf := proto.EncodeVarint(bndry)
	key := fmt.Sprintf("%s:%s:%s:last", msg.DataTuple.Context, msg.DataTuple.Subject, msg.DataTuple.Predicate)
	err = db.Put([]byte(key), buf, nil)
	if err != nil {
		return errors.Wrap(err, "unable to save allocated tuple boundary number")
	}

	// update the message
	msg.DataTuple.Version = bndry

	return nil
}

//
// returns the boundary number for the last known instance of the
// data tuple in the message
//
func (bs *BoundaryService) GetLast(msg *messages.N3Message) (uint64, error) {

	db := bs.client.db
	var last uint64

	// create the db lookup key from the message & retrieve
	key := fmt.Sprintf("%s:%s:%s:last", msg.DataTuple.Context, msg.DataTuple.Subject, msg.DataTuple.Predicate)
	// log.Println("boundary key: ", key)
	val, err := db.Get([]byte(key), nil)
	if err != nil && err != leveldb.ErrNotFound {
		log.Println("error on boundary last read: ", err)
		return last, err
	}

	last, _ = proto.DecodeVarint(val)

	return last, nil

}

//
// close the boundary datastore
//
func (bs *BoundaryService) Close() error {
	err := bs.client.closeStore()
	if err != nil {
		return err
	}
	return nil
}
