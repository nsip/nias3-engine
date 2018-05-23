// hexastore-service.go

package leveldb

import (
	"fmt"

	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type HexastoreService struct {
	client *hexastoreClient
}

// Ensure HexastoreService implements n3.HexastoreService interface
var _ n3.HexastoreService = &HexastoreService{}

func NewHexastoreService() (*HexastoreService, error) {

	hs := &HexastoreService{}
	client, err := newHexastoreClient()
	if err != nil {
		return nil, err
	}
	hs.client = client

	return hs, nil

}

//
// returns the currently stored value (O:property) of the provided data tuple
//
func (hs *HexastoreService) GetCurrentValue(dt *messages.DataTuple) (string, error) {

	db := hs.client.db

	// create the db lookup key from the tuple & retrieve
	key := fmt.Sprintf("%s:%s:%s:", dt.Context, dt.Subject, dt.Predicate)

	var val string
	iter := db.NewIterator(util.BytesPrefix([]byte(key)), nil)
	for iter.Next() {
		val = fmt.Sprintf("%s", iter.Value())
	}
	iter.Release()
	err := iter.Error()

	// val, err := db.Get([]byte(key), nil)
	if err != nil && err != leveldb.ErrNotFound {
		return val, err
	}

	return val, nil

}

//
// commits the data tuple to the hexastore & inflates to all permutations
//
func (hs *HexastoreService) Commit(msg *messages.N3Message) error {

	db := hs.client.db

	// commit the basic tuple with its object value
	dt := msg.DataTuple
	key := fmt.Sprintf("%s:%s:%s:%s", dt.Context, dt.Subject, dt.Predicate, dt.Object)
	err := db.Put([]byte(key), []byte(dt.Object), nil)
	if err != nil {
		return err
	}

	// log.Printf("successfully committed:\n\tkey:%s\n\tvalue:%s", key, dt.Object)

	return nil
}

//
// close the hexa datastore
//
func (hs *HexastoreService) Close() error {
	err := hs.client.closeStore()
	if err != nil {
		return err
	}
	return nil
}
