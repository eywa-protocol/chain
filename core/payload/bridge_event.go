package payload

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
	"github.com/near/borsh-go"
)

type BridgeEventData struct {
	RequestType    string
	Bridge         ethcommon.Address
	ReqId          RequestId
	Selector       []byte
	ReceiveSide    ethcommon.Address
	OppositeBridge ethcommon.Address
	ChainId        uint64
}

type BridgeEvent struct {
	data   BridgeEventData
	txHash []byte
}

func NewBridgeEvent(data *wrappers.BridgeOracleRequest) *BridgeEvent {
	return &BridgeEvent{
		data: BridgeEventData{
			RequestType:    data.RequestType,
			Bridge:         data.Bridge,
			ReqId:          data.RequestId,
			Selector:       data.Selector,
			ReceiveSide:    data.ReceiveSide,
			OppositeBridge: data.OppositeBridge,
			ChainId:        data.ChainId.Uint64(),
		},
		txHash: data.Raw.TxHash[:],
	}
}

func (e *BridgeEvent) TxType() TransactionType {
	return BridgeEventType
}

func (e *BridgeEvent) RequestState() RequestState {
	return ReqStateReceived
}

func (e *BridgeEvent) RequestId() RequestId {
	return e.data.ReqId
}

func (e *BridgeEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e.data)
}

func (e *BridgeEvent) SrcTxHash() []byte {
	return e.txHash
}

func (e *BridgeEvent) DstChainId() (uint64, bool) {
	return e.data.ChainId, false
}

func (e *BridgeEvent) Data() interface{} {
	return e.data
}

func (e *BridgeEvent) Deserialization(source *common.ZeroCopySource) error {
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

func (e *BridgeEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := MarshalBinary(&e.data)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func MarshalBinary(be *BridgeEventData) (data []byte, err error) {
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

func (e *BridgeEvent) RawData() []byte {
	// Must be binary compartible with SolanaToEVMEvent
	var bridgeFrom [32]byte
	copy(bridgeFrom[:], e.data.Bridge[:])

	sink := common.NewZeroCopySink(nil)
	sink.WriteBytes(e.data.ReqId[:])       // 32 bytes
	sink.WriteBytes(bridgeFrom[:])         // 32 bytes as in SolanaToEvmEvent
	sink.WriteBytes(e.data.ReceiveSide[:]) // 20 bytes
	sink.WriteVarBytes(e.data.Selector)
	sink.WriteUint64(e.data.ChainId)
	return sink.Bytes()
}
