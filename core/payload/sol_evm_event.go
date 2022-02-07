package payload

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/eywa-protocol/chain/common"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
	"math/big"
)

type SolanaToEVMEvent struct {
	OriginData bridge.BridgeEvent
}

// `ContractInvokeParam.Args` has reference of `source`
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

func (self *SolanaToEVMEvent) Serialization(sink *common.ZeroCopySink) {
	oracleRequestBytes, _ := MarshalBinarySolanaToEVMEvent(&self.OriginData)
	sink.WriteVarBytes(oracleRequestBytes)
}

func unmarshalBinarySolanaToEVMEvent(data []byte, st *bridge.BridgeEvent) error {
	r := bytes.NewReader(data)
	var dec struct {
		RequestType    string
		Bridge         [32]byte
		RequestId      [32]byte
		Selector       []byte
		ReceiveSide    [32]byte
		OppositeBridge [20]byte
		Chainid        *big.Int
		LogResult      ws.LogResult
	}
	gob.NewDecoder(r).Decode(&dec)
	st.RequestType = dec.RequestType

	st.BridgePubKey = dec.Bridge
	st.RequestId = dec.RequestId
	st.Selector = dec.Selector
	st.OppositeBridge = dec.OppositeBridge
	st.ChainId = dec.Chainid.Uint64()
	st.LogResult = dec.LogResult

	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func MarshalBinarySolanaToEVMEvent(be *bridge.BridgeEvent) (data []byte, err error) {
	var (
		b bytes.Buffer
		w = bufio.NewWriter(&b)
	)
	if err := gob.NewEncoder(w).Encode(struct {
		RequestType    string
		Bridge         [32]byte
		RequestId      [32]byte
		Selector       []byte
		OppositeBridge [20]byte
		Chainid        *big.Int
		LogResult      ws.LogResult
	}{
		be.RequestType,
		be.BridgePubKey,
		be.RequestId,
		be.Selector,
		be.OppositeBridge,
		big.NewInt(int64(be.ChainId)),
		be.LogResult,
	}); err != nil {
		return nil, err
	}

	w.Flush()
	return b.Bytes(), nil
}
