package payload

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/near/borsh-go"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"

	"github.com/eywa-protocol/chain/common"
)

type SolanaToEVMEvent struct {
	RequestType    string
	BridgePubKey   Bytes32
	ReqId          RequestId
	Selector       []byte
	ReceiveSide    ethcommon.Address
	OppositeBridge ethcommon.Address
	ChainId        uint64
	Signature      Bytes64
	Slot           uint64
}

func NewSolanaToEVMEvent(data *bridge.BridgeEvent) *SolanaToEVMEvent {
	return &SolanaToEVMEvent{
		RequestType:    data.RequestType,
		BridgePubKey:   Bytes32(data.BridgePubKey),
		ReqId:          RequestId(data.RequestId),
		Selector:       data.Selector,
		ReceiveSide:    data.ReceiveSide,
		OppositeBridge: data.OppositeBridge,
		ChainId:        data.ChainId,
		Signature:      Bytes64(data.Signature),
		Slot:           data.Slot,
	}
}

func (e *SolanaToEVMEvent) TxType() TransactionType {
	return SolanaToEVMEventType
}

func (e *SolanaToEVMEvent) RequestState() RequestState {
	return ReqStateReceived
}

func (e *SolanaToEVMEvent) RequestId() RequestId {
	return e.ReqId
}

func (e *SolanaToEVMEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e)
}

func (e *SolanaToEVMEvent) SrcTxHash() []byte {
	return e.Signature[:]
}

func (e *SolanaToEVMEvent) DstChainId() (uint64, bool) {
	return e.ChainId, false
}

func (e *SolanaToEVMEvent) Data() interface{} {
	return e
}

func (e *SolanaToEVMEvent) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}
	err := borsh.Deserialize(e, code)
	if err != nil {
		return err
	}
	return nil
}

func (e *SolanaToEVMEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := marshalBinarySolanaToEVMEvent(e)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

// marshalBinarySolanaToEVMEvent MarshalBinary implements encoding.BinaryMarshaler
func marshalBinarySolanaToEVMEvent(be *SolanaToEVMEvent) (data []byte, err error) {
	var (
		b bytes.Buffer
		w = bufio.NewWriter(&b)
	)

	br := *be
	if err := borsh.NewEncoder(w).Encode(br); err != nil {
		return nil, err
	}

	if err := w.Flush(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil

}

func (e *SolanaToEVMEvent) RawData() []byte {
	// Must be binary compartible with BridgeEvent
	sink := common.NewZeroCopySink(nil)
	sink.WriteBytes(e.ReqId[:])        // 32 bytes
	sink.WriteBytes(e.BridgePubKey[:]) // 32 bytes
	sink.WriteBytes(e.ReceiveSide[:])  // 20 bytes
	sink.WriteVarBytes(e.Selector)
	sink.WriteUint64(e.ChainId)
	return sink.Bytes()
}
