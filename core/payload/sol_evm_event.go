package payload

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/near/borsh-go"

	"github.com/eywa-protocol/chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
)

type SolanaToEVMEvent struct {
	OriginData bridge.BridgeEvent
}

func (e *SolanaToEVMEvent) TxType() TransactionType {
	return SolanaToEVMEventType
}

func (e *SolanaToEVMEvent) RequestState() RequestState {
	return ReqStateReceived
}

func (e *SolanaToEVMEvent) RequestId() RequestId {
	return RequestId(e.OriginData.RequestId)
}

func (e *SolanaToEVMEvent) ToJson() (json.RawMessage, error) {
	return json.Marshal(e.OriginData)
}

func (e *SolanaToEVMEvent) SrcTxHash() []byte {
	return e.OriginData.Signature[:]
}

func (e *SolanaToEVMEvent) DstChainId() (uint64, bool) {
	return e.OriginData.ChainId, false
}

func (e *SolanaToEVMEvent) Deserialization(source *common.ZeroCopySource) error {
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

func (e *SolanaToEVMEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := marshalBinarySolanaToEVMEvent(&e.OriginData)
	if err != nil {
		return err
	}
	sink.WriteVarBytes(oracleRequestBytes)
	return nil
}

// marshalBinarySolanaToEVMEvent MarshalBinary implements encoding.BinaryMarshaler
func marshalBinarySolanaToEVMEvent(be *bridge.BridgeEvent) (data []byte, err error) {
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
	sink.WriteBytes(e.OriginData.RequestId[:])    // 32 bytes
	sink.WriteBytes(e.OriginData.BridgePubKey[:]) // 32 bytes
	sink.WriteBytes(e.OriginData.ReceiveSide[:])  // 20 bytes
	sink.WriteVarBytes(e.OriginData.Selector)
	sink.WriteUint64(e.OriginData.ChainId)
	return sink.Bytes()
}
