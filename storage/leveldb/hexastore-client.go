// hexastore-client.go

package leveldb

import (
	"log"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

var hexastoreDB *leveldb.DB

type hexastoreClient struct {
	path string
	db   *leveldb.DB
}

func newHexastoreClient() (*hexastoreClient, error) {
	hc := &hexastoreClient{path: "./db/hexastore"}
	err := hc.open()
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to hexastore db")
	}
	return hc, nil
}

// accesses instance of the k/v boundary data-store.
func (hc *hexastoreClient) open() error {
	if hexastoreDB == nil {
		var dbErr error
		hexastoreDB, dbErr = leveldb.OpenFile(hc.path, nil)
		if dbErr != nil {
			return dbErr
		}
	}
	hc.db = boundaryDB
	log.Println("client connected to hexastore db")

	return nil
}

// Closes the underlying data-store
func (hc *hexastoreClient) closeStore() error {
	if hc.db != nil {
		log.Println("hexastore db closed")
		return hc.db.Close()
	}
	return nil
}
