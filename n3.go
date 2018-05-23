// n3.go

package n3

import "github.com/nsip/nias3-engine/messages"

const N3ClientVersion = "n3-p2p-node/0.0.1"

//
//
// interfaces
//

//
// storage layer
//

type LedgerService interface {
	CommitNew(*messages.N3Message) error
	Append(*messages.N3Message) error
	AssignSequence(*messages.N3Message) error
	CreateSyncDigest(context string) (map[string]uint64, error)
	Close() error
}

type BoundaryService interface {
	AssignBoundary(*messages.N3Message) error
	GetLast(*messages.N3Message) (uint64, error)
	Close() error
}

type HexastoreService interface {
	Commit(*messages.N3Message) error
	GetCurrentValue(*messages.DataTuple) (string, error)
	Close() error
}

//
// streaming layer
//

type StreamClient interface {
	ReaderService() ReaderService
	WriterService() WriterService
}

type ReaderService interface {
	Subscribe(topic string, clientID string) error
	MsgC() <-chan *messages.SignedN3Message
	ErrC() <-chan error
	Close() error
}

type WriterService interface {
	Subscribe(topic string, clientID string) error
	MsgC() chan<- *messages.SignedN3Message
	ErrC() <-chan error
	Close() error
}

//
// crypto
//

type CryptoService interface {
	SignMessage(msg *messages.N3Message) (*messages.SignedN3Message, error)
	//param: gossip allows nodes to propagate messages autonomously: defaults to false
	NewMessageData(gossip bool) (*messages.MessageData, error)
	AuthenticateMessage(msg *messages.SignedN3Message) (bool, error)
	PublicID() (string, error)
}

//
// trust
//
