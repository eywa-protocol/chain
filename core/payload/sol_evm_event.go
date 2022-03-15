package payload

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/gagliardetto/solana-go"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
)

type SolanaToEVMEvent struct {
	OriginData bridge.BridgeEvent
}

func (self *SolanaToEVMEvent) TxType() TransactionType {
	return SolanaToEVMEventType
}

func (self *SolanaToEVMEvent) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}
	err := unmarshalBinarySolanaToEVMEvent(code, &self.OriginData)
	if err != nil {
		return err
	}
	return nil
}

func (self *SolanaToEVMEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := MarshalBinarySolanaToEVMEvent(&self.OriginData)
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
	gob.NewDecoder(r).Decode(&dec)
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

// MarshalBinary implements encoding.BinaryMarshaler
func MarshalBinarySolanaToEVMEvent(be *bridge.BridgeEvent) (data []byte, err error) {

	var (
		b bytes.Buffer
		w = bufio.NewWriter(&b)
	)

	br := be
	if err := gob.NewEncoder(w).Encode(br); err != nil {
		return nil, err
	}

	w.Flush()
	return b.Bytes(), nil
}

func (self *SolanaToEVMEvent) RawData() []byte {
	var data []byte
	data = append(data, self.OriginData.RequestId[:]...)
	data = append(data, self.OriginData.Selector...)
	data = append(data, self.OriginData.ReceiveSide[:]...)
	data = append(data, self.OriginData.BridgePubKey[:]...)
	return data
}
