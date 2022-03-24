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

type BridgeEvent struct {
	OriginData wrappers.BridgeOracleRequest
}

func (e *BridgeEvent) TxType() TransactionType {
	return BridgeEventType
}

func (e *BridgeEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e.OriginData)
}

func (e *BridgeEvent) DstChainId() (uint64, bool) {

	return e.OriginData.Chainid.Uint64(), false
}

func (e *BridgeEvent) Deserialization(source *common.ZeroCopySource) error {
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

func (e *BridgeEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := MarshalBinary(&e.OriginData)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func MarshalBinary(be *wrappers.BridgeOracleRequest) (data []byte, err error) {
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
	sink := common.NewZeroCopySink(nil)
	sink.WriteBytes(e.OriginData.Bridge[:])
	sink.WriteBytes(e.OriginData.RequestId[:])
	sink.WriteVarBytes(e.OriginData.Selector)
	sink.WriteBytes(e.OriginData.ReceiveSide[:])
	sink.WriteUint64(e.OriginData.Chainid.Uint64())
	return sink.Bytes()
}
