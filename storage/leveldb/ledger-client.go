package leveldb

import (
	"log"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

var ledgerDB *leveldb.DB

type ledgerClient struct {
	path string
	db   *leveldb.DB
}

func newLedgerClient() (*ledgerClient, error) {
	lc := &ledgerClient{path: "./db/ledger"}
	err := lc.open()
	if err != nil {
		return nil, errors.Wrap(err, "unable to open ledger store")
	}
	return lc, nil
}

// accesses instance of the k/v ledger data-store.
func (lc *ledgerClient) open() error {
	if ledgerDB == nil {
		var dbErr error
		ledgerDB, dbErr = leveldb.OpenFile(lc.path, nil)
		if dbErr != nil {
			return dbErr
		}
	}
	lc.db = ledgerDB
	// log.Println("client connected to ledger db")

	return nil
}

// Closes the underlying data-store
func (lc *ledgerClient) closeStore() error {
	if lc.db != nil {
		log.Println("ledger db closed")
		return lc.db.Close()
	}
	return nil
}
