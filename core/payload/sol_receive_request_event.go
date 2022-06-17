package payload

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/near/borsh-go"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
)

type SolReceiveRequestEvent struct {
	OriginData bridge.BridgeReceiveEvent
}

func (e *SolReceiveRequestEvent) TxType() TransactionType {
	return SolReceiveRequestEventType
}

func (e *SolReceiveRequestEvent) RequestState() RequestState {
	return ReqStateSent
}

func (e *SolReceiveRequestEvent) RequestId() RequestId {
	return RequestId(e.OriginData.RequestId)
}

func (e *SolReceiveRequestEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e.OriginData)
}

func (e *SolReceiveRequestEvent) SrcTxHash() []byte {
	return e.OriginData.Signature[:]
}

func (e *SolReceiveRequestEvent) DstChainId() (uint64, bool) {
	return 0, true
}

func (e *SolReceiveRequestEvent) Deserialization(source *common.ZeroCopySource) error {
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

func (e *SolReceiveRequestEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := marshalBinarySolReceiveRequest(&e.OriginData)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

func marshalBinarySolReceiveRequest(be *bridge.BridgeReceiveEvent) (data []byte, err error) {
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

func (e *SolReceiveRequestEvent) RawData() []byte {
	var data []byte
	data = append(data, e.OriginData.RequestId[:]...)
	return data
}
