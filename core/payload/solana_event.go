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

type BridgeSolanaEvent struct {
	OriginData wrappers.BridgeOracleRequestSolana
}

func (e *BridgeSolanaEvent) TxType() TransactionType {
	return BridgeEventSolanaType
}

func (e *BridgeSolanaEvent) RequestState() RequestState {
	return ReqStateReceived
}

func (e *BridgeSolanaEvent) RequestId() RequestId {
	return RequestId(e.OriginData.RequestId)
}

func (e *BridgeSolanaEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e.OriginData)
}

func (e *BridgeSolanaEvent) SrcTxHash() []byte {
	return e.OriginData.Raw.TxHash[:]
}

func (e *BridgeSolanaEvent) DstChainId() (uint64, bool) {
	return e.OriginData.ChainId.Uint64(), false
}

func (e *BridgeSolanaEvent) Data() interface{} {
	return e.OriginData
}

func (e *BridgeSolanaEvent) Deserialization(source *common.ZeroCopySource) error {
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

func (e *BridgeSolanaEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := marshalSolBinary(&e.OriginData)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

// marshalSolBinary MarshalBinary implements encoding.BinaryMarshaler
func marshalSolBinary(be *wrappers.BridgeOracleRequestSolana) (data []byte, err error) {
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
	sink.WriteBytes(e.OriginData.RequestId[:])      // 32 bytes
	sink.WriteBytes(e.OriginData.Bridge[:])         // 32 bytes
	sink.WriteBytes(e.OriginData.OppositeBridge[:]) // 32 bytes
	sink.WriteVarBytes(e.OriginData.Selector)
	sink.WriteUint64(e.OriginData.ChainId.Uint64())
	return sink.Bytes()
}
