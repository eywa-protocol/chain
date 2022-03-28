package payload

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/gagliardetto/solana-go"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
)

type SolanaToEVMEvent struct {
	OriginData bridge.BridgeEvent
}

func (e *SolanaToEVMEvent) TxType() TransactionType {
	return SolanaToEVMEventType
}

func (e *SolanaToEVMEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e.OriginData)
}

func (e *SolanaToEVMEvent) DstChainId() (uint64, bool) {
	return e.OriginData.ChainId, false
}

func (e *SolanaToEVMEvent) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}
	err := unmarshalBinarySolanaToEVMEvent(code, &e.OriginData)
	if err != nil {
		return err
	}
	return nil
}

func (e *SolanaToEVMEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := MarshalBinarySolanaToEVMEvent(&e.OriginData)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

func unmarshalBinarySolanaToEVMEvent(data []byte, st *bridge.BridgeEvent) error {
	r := bytes.NewReader(data)
	var dec = bridge.BridgeEvent{
		OracleRequest: bridge.OracleRequest{},
		Signature:     solana.Signature{},
		Slot:          0,
	}
	if err := gob.NewDecoder(r).Decode(&dec); err != nil {
		return err
	}
	st.ChainId = dec.ChainId
	st.BridgePubKey = dec.BridgePubKey
	st.Slot = dec.Slot
	st.OppositeBridge = dec.OppositeBridge
	st.ReceiveSide = dec.ReceiveSide
	st.Signature = dec.Signature
	st.RequestType = dec.RequestType
	st.Selector = dec.Selector
	st.RequestId = dec.RequestId

	return nil
}

// MarshalBinarySolanaToEVMEvent MarshalBinary implements encoding.BinaryMarshaler
func MarshalBinarySolanaToEVMEvent(be *bridge.BridgeEvent) (data []byte, err error) {
	var (
		b bytes.Buffer
		w = bufio.NewWriter(&b)
	)

	br := be
	if err := gob.NewEncoder(w).Encode(br); err != nil {
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
	sink.WriteBytes(e.OriginData.RequestId[:])    // 32 bytes
	sink.WriteBytes(e.OriginData.BridgePubKey[:]) // 32 bytes
	sink.WriteBytes(e.OriginData.ReceiveSide[:])  // 20 bytes
	sink.WriteVarBytes(e.OriginData.Selector)
	sink.WriteUint64(e.OriginData.ChainId)
	return sink.Bytes()
}
