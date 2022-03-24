package payload

import (
	"encoding/json"

	"github.com/eywa-protocol/chain/common"
)

type TransactionType byte

const (
	InvokeType              TransactionType = 0xd1
	NodeType                TransactionType = 0xd2
	EpochType               TransactionType = 0x22
	UpTimeType              TransactionType = 0xd4
	BridgeEventType         TransactionType = 0x1f
	BridgeEventSolanaType   TransactionType = 0x20
	SolanaToEVMEventType    TransactionType = 0x21
	ReceiveRequestEventType TransactionType = 0x23
)

func (tt TransactionType) String() string {
	switch tt {
	case InvokeType:
		return "invoke"
	case NodeType:
		return "node"
	case EpochType:
		return "epoch"
	case UpTimeType:
		return "up_time"
	case BridgeEventType:
		return "bridge_event"
	case ReceiveRequestEventType:
		return "receive_request_event"
	default:
		return "unknown"
	}
}

type Payload interface {
	TxType() TransactionType
	ToJson() (json.RawMessage, error)
	DstChainId() (uint64, bool)
	Serialization(*common.ZeroCopySink) error
	Deserialization(*common.ZeroCopySource) error
	RawData() []byte
}
