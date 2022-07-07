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

type SolanaToEVMEventData struct {
	RequestType    string
	BridgePubKey   Bytes32
	RequestId      RequestId
	Selector       []byte
	ReceiveSide    ethcommon.Address
	OppositeBridge ethcommon.Address
	ChainId        uint64
	Signature      Bytes64
	Slot           uint64
}

type SolanaToEVMEvent struct {
	data SolanaToEVMEventData
}

func NewSolanaToEVMEvent(data *bridge.BridgeEvent) *SolanaToEVMEvent {
	return &SolanaToEVMEvent{
		data: SolanaToEVMEventData{
			RequestType:    data.RequestType,
			BridgePubKey:   Bytes32(data.BridgePubKey),
			RequestId:      RequestId(data.RequestId),
			Selector:       data.Selector,
			ReceiveSide:    data.ReceiveSide,
			OppositeBridge: data.OppositeBridge,
			ChainId:        data.ChainId,
			Signature:      Bytes64(data.Signature),
			Slot:           data.Slot,
		},
	}
}

func (e *SolanaToEVMEvent) TxType() TransactionType {
	return SolanaToEVMEventType
}

func (e *SolanaToEVMEvent) RequestState() RequestState {
	return ReqStateReceived
}

func (e *SolanaToEVMEvent) RequestId() RequestId {
	return RequestId(e.data.RequestId)
}

func (e *SolanaToEVMEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e.data)
}

func (e *SolanaToEVMEvent) SrcTxHash() []byte {
	return e.data.Signature[:]
}

func (e *SolanaToEVMEvent) DstChainId() (uint64, bool) {
	return e.data.ChainId, false
}

func (e *SolanaToEVMEvent) Data() interface{} {
	return e.data
}

func (e *SolanaToEVMEvent) Deserialization(source *common.ZeroCopySource) error {
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

func (e *SolanaToEVMEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := marshalBinarySolanaToEVMEvent(&e.data)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

// marshalBinarySolanaToEVMEvent MarshalBinary implements encoding.BinaryMarshaler
func marshalBinarySolanaToEVMEvent(be *SolanaToEVMEventData) (data []byte, err error) {
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
	sink.WriteBytes(e.data.RequestId[:])    // 32 bytes
	sink.WriteBytes(e.data.BridgePubKey[:]) // 32 bytes
	sink.WriteBytes(e.data.ReceiveSide[:])  // 20 bytes
	sink.WriteVarBytes(e.data.Selector)
	sink.WriteUint64(e.data.ChainId)
	return sink.Bytes()
}
