package payload

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/eywa-protocol/wrappers"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"math/big"
)

type BridgeEvent struct {
	OriginData wrappers.BridgeOracleRequest
}

// `ContractInvokeParam.Args` has reference of `source`
func (self *BridgeEvent) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}
	err := unmarshalBinary(code, &self.OriginData)
	if err != nil {
		return err
	}
	return nil
}

func (self *BridgeEvent) Serialization(sink *common.ZeroCopySink) {
	oracleRequestBytes, _ := MarshalBinary(&self.OriginData)
	sink.WriteVarBytes(oracleRequestBytes)
}

func unmarshalBinary(data []byte, st *wrappers.BridgeOracleRequest) error {
	r := bytes.NewReader(data)
	var dec struct {
		RequestType    string
		Bridge         ethCommon.Address
		RequestId      [32]byte
		Selector       []byte
		ReceiveSide    ethCommon.Address
		OppositeBridge ethCommon.Address
		Chainid        *big.Int
		Raw            types.Log
	}
	gob.NewDecoder(r).Decode(&dec)
	st.RequestType = dec.RequestType
	st.Bridge = dec.Bridge
	st.RequestId = dec.RequestId
	st.Selector = dec.Selector
	st.ReceiveSide = dec.ReceiveSide
	st.OppositeBridge = dec.OppositeBridge
	st.Chainid = dec.Chainid
	st.Raw = dec.Raw

	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func MarshalBinary(be *wrappers.BridgeOracleRequest) (data []byte, err error) {
	var (
		b bytes.Buffer
		w = bufio.NewWriter(&b)
	)
	if err := gob.NewEncoder(w).Encode(struct {
		RequestType    string
		Bridge         ethCommon.Address
		RequestId      [32]byte
		Selector       []byte
		ReceiveSide    ethCommon.Address
		OppositeBridge ethCommon.Address
		Chainid        *big.Int
		Raw            types.Log
	}{
		be.RequestType,
		be.Bridge,
		be.RequestId,
		be.Selector,
		be.ReceiveSide,
		be.OppositeBridge,
		be.Chainid,
		be.Raw,
	}); err != nil {
		return nil, err
	}

	w.Flush()
	return b.Bytes(), nil
}
