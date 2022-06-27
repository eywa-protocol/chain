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

type ReceiveRequestEvent struct {
	OriginData wrappers.BridgeReceiveRequest
}

func (e *ReceiveRequestEvent) TxType() TransactionType {
	return ReceiveRequestEventType
}

func (e *ReceiveRequestEvent) RequestState() RequestState {
	return ReqStateSent
}

func (e *ReceiveRequestEvent) RequestId() RequestId {
	return e.OriginData.ReqId
}

func (e *ReceiveRequestEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e.OriginData)
}

func (e *ReceiveRequestEvent) SrcTxHash() []byte {
	return e.OriginData.Raw.TxHash[:]
}

func (e *ReceiveRequestEvent) DstChainId() (uint64, bool) {
	return 0, true
}

func (e *ReceiveRequestEvent) Data() interface{} {
	return e.OriginData
}

func (e *ReceiveRequestEvent) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}
	err := borsh.Deserialize(&e.OriginData, code)
	if err != nil {
		return err
	}
	return nil
}

func (e *ReceiveRequestEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := marshalBinaryRecievRequest(&e.OriginData)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

func marshalBinaryRecievRequest(be *wrappers.BridgeReceiveRequest) (data []byte, err error) {
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
	data = append(data, e.OriginData.ReqId[:]...)
	return data
}
