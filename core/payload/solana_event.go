package payload

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
)

type BridgeSolanaEvent struct {
	OriginData wrappers.BridgeOracleRequestSolana
}

func (e *BridgeSolanaEvent) TxType() TransactionType {
	return BridgeEventSolanaType
}

func (e *BridgeSolanaEvent) RequestState() ReqState {
	return ReqStateReceived
}

func (e *BridgeSolanaEvent) RequestId() [32]byte {
	return e.OriginData.RequestId
}

func (e *BridgeSolanaEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e.OriginData)
}

func (e *BridgeSolanaEvent) DstChainId() (uint64, bool) {
	return e.OriginData.ChainId.Uint64(), false
}

func (e *BridgeSolanaEvent) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}
	err := unmarshalBinarySolana(code, &e.OriginData)
	if err != nil {
		return err
	}
	return nil
}

func (e *BridgeSolanaEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := MarshalSolBinary(&e.OriginData)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

func unmarshalBinarySolana(data []byte, st *wrappers.BridgeOracleRequestSolana) error {
	r := bytes.NewReader(data)
	var dec struct {
		RequestType    string
		Bridge         [32]byte
		RequestId      [32]byte
		Selector       []byte
		ReceiveSide    [32]byte
		OppositeBridge [32]byte
		ChainId        *big.Int
		Raw            types.Log
	}
	if err := gob.NewDecoder(r).Decode(&dec); err != nil {
		return err
	}
	st.RequestType = dec.RequestType

	st.Bridge = dec.Bridge
	st.RequestId = dec.RequestId
	st.Selector = dec.Selector
	st.OppositeBridge = dec.OppositeBridge
	st.ChainId = dec.ChainId
	st.Raw = dec.Raw

	return nil
}

// MarshalSolBinary MarshalBinary implements encoding.BinaryMarshaler
func MarshalSolBinary(be *wrappers.BridgeOracleRequestSolana) (data []byte, err error) {
	var (
		b bytes.Buffer
		w = bufio.NewWriter(&b)
	)
	if err := gob.NewEncoder(w).Encode(struct {
		RequestType    string
		Bridge         [32]byte
		RequestId      [32]byte
		Selector       []byte
		OppositeBridge [32]byte
		ChainId        *big.Int
		Raw            types.Log
	}{
		be.RequestType,
		be.Bridge,
		be.RequestId,
		be.Selector,
		be.OppositeBridge,
		be.ChainId,
		be.Raw,
	}); err != nil {
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
