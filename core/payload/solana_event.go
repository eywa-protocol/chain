package payload

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
	"github.com/near/borsh-go"
)

type BridgeSolanaEventData struct {
	RequestType    string
	Bridge         Bytes32
	ReqId          RequestId
	Selector       []byte
	OppositeBridge Bytes32
	ChainId        uint64
}

type BridgeSolanaEvent struct {
	data   BridgeSolanaEventData
	txHash []byte
}

func NewBridgeSolanaEvent(data *wrappers.BridgeOracleRequestSolana) *BridgeSolanaEvent {
	return &BridgeSolanaEvent{
		data: BridgeSolanaEventData{
			RequestType:    data.RequestType,
			Bridge:         data.Bridge,
			ReqId:          RequestId(data.RequestId),
			Selector:       data.Selector,
			OppositeBridge: data.OppositeBridge,
			ChainId:        data.ChainId.Uint64(),
		},
		txHash: data.Raw.TxHash[:],
	}
}

func (e *BridgeSolanaEvent) TxType() TransactionType {
	return BridgeEventSolanaType
}

func (e *BridgeSolanaEvent) RequestState() RequestState {
	return ReqStateReceived
}

func (e *BridgeSolanaEvent) RequestId() RequestId {
	return e.data.ReqId
}

func (e *BridgeSolanaEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e.data)
}

func (e *BridgeSolanaEvent) SrcTxHash() []byte {
	return e.txHash[:]
}

func (e *BridgeSolanaEvent) DstChainId() (uint64, bool) {
	return e.data.ChainId, false
}

func (e *BridgeSolanaEvent) Data() interface{} {
	return e.data
}

func (e *BridgeSolanaEvent) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}
	err := borsh.Deserialize(&e.data, code)
	if err != nil {
		return err
	}
	return nil
}

func (e *BridgeSolanaEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := marshalSolBinary(&e.data)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

// marshalSolBinary MarshalBinary implements encoding.BinaryMarshaler
func marshalSolBinary(be *BridgeSolanaEventData) (data []byte, err error) {
	var (
		b bytes.Buffer
		w = bufio.NewWriter(&b)
	)
	qwf := *be
	if err := borsh.NewEncoder(w).Encode(qwf); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *BridgeSolanaEvent) RawData() []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteBytes(e.data.ReqId[:])          // 32 bytes
	sink.WriteBytes(e.data.Bridge[:])         // 32 bytes
	sink.WriteBytes(e.data.OppositeBridge[:]) // 32 bytes
	sink.WriteVarBytes(e.data.Selector)
	sink.WriteUint64(e.data.ChainId)
	return sink.Bytes()
}
