// ledger-service.go

package leveldb

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/nsip/nias3-engine"
	"github.com/nsip/nias3-engine/messages"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	ErrOutOfSequnce    = errors.New("message out of sequence")
	ErrTamperedMessage = errors.New("message failed blockchain validation")
	ErrCommitFail      = errors.New("could not commit to ledger db")
)

type LedgerService struct {
	client *ledgerClient
}

// Ensure LedgerService implements n3.LedgerService interface
var _ n3.LedgerService = &LedgerService{}

func NewLedgerService() (*LedgerService, error) {

	ls := &LedgerService{}
	client, err := newLedgerClient()
	if err != nil {
		return nil, err
	}
	ls.client = client

	return ls, nil

}

//
// finds the next available blockchain sequnce number for this
// user and assigns it to the given message
//
func (ls *LedgerService) AssignSequence(msg *messages.N3Message) error {

	var seq uint64
	db := ls.client.db

	// create the db lookup key from the message & retrieve
	key := fmt.Sprintf("last:%s:%s", msg.DataTuple.Context, msg.MessageData.NodeId)
	// log.Println("sequnce key: ", key)
	val, err := db.Get([]byte(key), nil)
	if err != nil && err != leveldb.ErrNotFound {
		log.Println("error on sequence read: ", err)
		return err
	}

	var last uint64
	last, _ = proto.DecodeVarint(val)
	// log.Println("decoded last: ", last)

	seq = last + 1
	// log.Println("assigned seq: ", seq)

	// write the new value for this user:context sequnce back to the db
	buf := proto.EncodeVarint(seq)
	err = db.Put([]byte(key), buf, nil)
	if err != nil {
		return errors.Wrap(err, "unable to save allocated sequence number")
	}

	// update the message
	msg.Sequence = seq

	// log.Println("sequence assigned: ", seq)

	return nil
}

//
// commits the necessary information to the ledger to confirm
// this message as part of the blockchain,typically called
// on an ingest activity where messages have no prior history
//
func (ls *LedgerService) CommitNew(msg *messages.N3Message) error {

	err := ls.assignPreviousHash(msg)
	if err != nil {
		return errors.Wrap(err, "commitnew: unable to establish previous b/c hash")
	}

	msgHash, err := hashMessage(msg)
	if err != nil {
		return errors.Wrap(err, "commitnew: unable to create message hash")
	}

	err = ls.ledgerCommit(msg, msgHash)
	if err != nil {
		return ErrCommitFail
	}

	msg.MsgHash = msgHash

	return nil
}

//
// commits messages that already contain their own blockchain information
// typically messages received from the replicate feed.
// checking here validates that no tampering has occurred, and if ok
// addds the blockchain information to the ledger.
//
func (ls *LedgerService) Append(msg *messages.N3Message) error {

	db := ls.client.db
	prevSeq := msg.Sequence - 1

	// check previous messages if available
	if prevSeq != 0 {

		// retrieve the previous message
		key := fmt.Sprintf("%s:%s:%d", msg.MessageData.NodeId, msg.DataTuple.Context, prevSeq)
		hashBytes, err := db.Get([]byte(key), nil)
		if err != nil {
			return ErrOutOfSequnce
		}
		hashString := fmt.Sprintf("%s", hashBytes)

		if msg.PrevHash != hashString {
			log.Printf("commit error: message id: %s has correct sequence but failed blockchain hash validation\n", msg.MessageData.Id)
			return ErrTamperedMessage
		}

	}

	// commit to the ledger
	err := ls.ledgerCommit(msg, msg.MsgHash)
	if err != nil {
		return ErrCommitFail
	}

	return nil
}

//
// creates a snapshot of the current ledger message sequnce by client
// for a given context.
// is used to construct a SyncRequest to a remote peer
// to ensure only new data is transmitted.
//
func (ls *LedgerService) CreateSyncDigest(context string) (map[string]uint64, error) {

	digest := make(map[string]uint64)
	db := ls.client.db

	prefix := fmt.Sprintf("last:%s:", context)
	iter := db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		key := fmt.Sprintf("%s", iter.Key())
		val, _ := proto.DecodeVarint(iter.Value())
		digest[key] = val
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, err
	}

	return digest, nil

}

//
// links this message to its predecessor in the blockchain
// by recording the previous messages's hash signature
//
func (ls *LedgerService) assignPreviousHash(msg *messages.N3Message) error {

	db := ls.client.db
	prevSeq := msg.Sequence - 1

	var hashBytes []byte
	var hashString string

	if prevSeq == 0 {
		// if no prev exists hash generate a new arbitrary 'genesis' hash
		// log.Println("creating random genesis block")
		buf := make([]byte, 32)
		_, err := rand.Read(buf)
		if err != nil {
			return errors.Wrap(err, "unable to create genesis hash")
		}
		hashBytes = buf
		hashString = fmt.Sprintf("%x", hashBytes)
	} else {
		key := fmt.Sprintf("%s:%s:%d", msg.MessageData.NodeId, msg.DataTuple.Context, prevSeq)
		hashBytes, err := db.Get([]byte(key), nil)
		if err != nil {
			return errors.Wrap(err, "unable to find previous b/c msg")
		}
		hashString = fmt.Sprintf("%s", hashBytes)
	}

	// log.Println("prev seq: ", prevSeq)

	msg.PrevHash = hashString

	return nil

}

//
// returns a sha256 hash checksum of a given message
//
func hashMessage(msg *messages.N3Message) (string, error) {

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return "", err
	}

	hashString := fmt.Sprintf("%x", sha256.Sum256(msgBytes))
	// log.Println("hash as string: ", hashString)

	return hashString, nil
}

//
// records the block with its sequence in the ledger
//
func (ls *LedgerService) ledgerCommit(msg *messages.N3Message, msgHash string) error {

	// add to the datastore
	db := ls.client.db
	key := fmt.Sprintf("%s:%s:%d", msg.MessageData.NodeId, msg.DataTuple.Context, msg.Sequence)
	err := db.Put([]byte(key), []byte(msgHash), nil)
	if err != nil {
		return err
	}

	return nil
}

//
// close the ledger datastore
//
func (ls *LedgerService) Close() error {
	err := ls.client.closeStore()
	if err != nil {
		return err
	}
	return nil
}
