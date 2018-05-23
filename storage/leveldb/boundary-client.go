package leveldb

import (
	"log"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

var boundaryDB *leveldb.DB

type boundaryClient struct {
	path string
	db   *leveldb.DB
}

func newBoundaryClient() (*boundaryClient, error) {
	bc := &boundaryClient{path: "./db/boundary"}
	err := bc.open()
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to boundary db")
	}
	return bc, nil
}

// accesses instance of the k/v boundary data-store.
func (bc *boundaryClient) open() error {
	if boundaryDB == nil {
		var dbErr error
		boundaryDB, dbErr = leveldb.OpenFile(bc.path, nil)
		if dbErr != nil {
			return dbErr
		}
	}
	bc.db = boundaryDB
	// log.Println("client connected to boundary db")

	return nil
}

// Closes the underlying data-store
func (bc *boundaryClient) closeStore() error {
	if bc.db != nil {
		log.Println("boundary db closed")
		return bc.db.Close()
	}
	return nil
}
