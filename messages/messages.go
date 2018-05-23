package messages

// DO NOT remove the line below, this is used by the
// go generate tool to create the protobuf message support classes
// for all p2p messages.
//
//go:generate protoc --go_out=. messages.proto
//

//
// all messages are defined in the protobuf (.proto) file which
// generates the required golang structs
//

import (
	"fmt"
	"log"

	"github.com/gogo/protobuf/proto"
)

//
// helper methods for messages
//

func (msg *N3Message) ToBytes() ([]byte, error) {

	// marshall data without the signature to protobufs3 binary format
	bin, err := proto.Marshal(msg)
	if err != nil {
		log.Println(err, "crpto:ToBytes failed to marshal message")
		return nil, err
	}
	return bin, nil
}

//
// generates a lookup key for this message to be used
// in comparing to a syncRequest digest
//
func (smsg *SignedN3Message) DigestKey() string {
	return fmt.Sprintf("last:%s:%s", smsg.N3Message.DataTuple.Context, smsg.N3Message.MessageData.NodeId)
}

func (smsg *SignedN3Message) Print() string {

	return fmt.Sprintf(`
		PrevHash: %s
		Sequence: %d
		MsgHash: %s
		Tuple: %s-%s-%s
		TupleVersion: %d
		Client: %s
		Timestamp: %d
		MsgId: %s
		NddeId: %s
		PubKey: %x
		`,
		// smsg.Signature,
		smsg.N3Message.PrevHash,
		smsg.N3Message.Sequence,
		smsg.N3Message.MsgHash,
		smsg.N3Message.DataTuple.GetSubject(),
		smsg.N3Message.DataTuple.GetObject(),
		smsg.N3Message.DataTuple.GetPredicate(),
		smsg.N3Message.DataTuple.GetVersion(),
		smsg.N3Message.MessageData.GetClientVersion(),
		smsg.N3Message.MessageData.GetTimestamp(),
		smsg.N3Message.MessageData.GetId(),
		smsg.N3Message.MessageData.GetNodeId(),
		smsg.N3Message.MessageData.GetNodePubKey())

}
