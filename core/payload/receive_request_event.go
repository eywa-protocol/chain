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

type ReceiveRequestEventData struct {
	ReqId       RequestId
	ReceiveSide ethcommon.Address
	BridgeFrom  Bytes32
}

type ReceiveRequestEvent struct {
	data   ReceiveRequestEventData
	txHash []byte
}

func NewReceiveRequestEvent(data *wrappers.BridgeReceiveRequest) *ReceiveRequestEvent {
	return &ReceiveRequestEvent{
		data: ReceiveRequestEventData{
			ReqId:       RequestId(data.ReqId),
			ReceiveSide: data.ReceiveSide,
			BridgeFrom:  data.BridgeFrom,
		},
		txHash: data.Raw.TxHash[:],
	}
}

func (e *ReceiveRequestEvent) TxType() TransactionType {
	return ReceiveRequestEventType
}

func (e *ReceiveRequestEvent) RequestState() RequestState {
	return ReqStateSent
}

func (e *ReceiveRequestEvent) RequestId() RequestId {
	return e.data.ReqId
}

func (e *ReceiveRequestEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e.data)
}

func (e *ReceiveRequestEvent) SrcTxHash() []byte {
	return e.txHash
}

func (e *ReceiveRequestEvent) DstChainId() (uint64, bool) {
	return 0, true
}

func (e *ReceiveRequestEvent) Data() interface{} {
	return e.data
}

func (e *ReceiveRequestEvent) Deserialization(source *common.ZeroCopySource) error {
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

func (e *ReceiveRequestEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := marshalBinaryRecievRequest(&e.data)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

func marshalBinaryRecievRequest(be *ReceiveRequestEventData) (data []byte, err error) {
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

func (e *ReceiveRequestEvent) RawData() []byte {
	var data []byte
	data = append(data, e.data.ReqId[:]...)
	return data
}
