package payload

import (
	"encoding/json"

	"github.com/eywa-protocol/chain/common"
)

type TransactionType byte

const (
	InvokeType                 TransactionType = 0xd1
	NodeType                   TransactionType = 0xd2
	EpochType                  TransactionType = 0x22
	UpTimeType                 TransactionType = 0xd4
	BridgeEventType            TransactionType = 0x1f
	BridgeEventSolanaType      TransactionType = 0x20
	SolanaToEVMEventType       TransactionType = 0x21
	ReceiveRequestEventType    TransactionType = 0x23
	SolReceiveRequestEventType TransactionType = 0x24
)

type ReqState uint8

const (
	ReqStateUnknown  ReqState = iota // request id not found in ledger
	ReqStateReceived                 // event received
	ReqStateSent                     // event sent to destination
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
	case BridgeEventSolanaType:
		return "bridge_event_solana"
	case SolanaToEVMEventType:
		return "solana_to_evm_event"
	case SolReceiveRequestEventType:
		return "solana_receive_request_event"
	default:
		return "unknown"
	}
}

type Payload interface {
	TxType() TransactionType
	RequestState() ReqState
	RequestId() [32]byte
	ToJson() (json.RawMessage, error)
	SrcTxHash() []byte
	DstChainId() (uint64, bool)
	Serialization(*common.ZeroCopySink) error
	Deserialization(*common.ZeroCopySource) error
	RawData() []byte
}
