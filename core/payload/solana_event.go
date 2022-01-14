package payload

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
)

type BridgeSolanaEvent struct {
	OriginData wrappers.BridgeOracleRequestSolana
}

// `ContractInvokeParam.Args` has reference of `source`
func (self *BridgeSolanaEvent) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}
	err := unmarshalSolBinary(code, &self.OriginData)
	if err != nil {
		return err
	}
	return nil
}

func (self *BridgeSolanaEvent) Serialization(sink *common.ZeroCopySink) {
	oracleRequestBytes, _ := MarshalSolBinary(&self.OriginData)
	sink.WriteVarBytes(oracleRequestBytes)
}

func unmarshalSolBinary(data []byte, st *wrappers.BridgeOracleRequestSolana) error {
	r := bytes.NewReader(data)
	var dec struct {
		RequestType    string
		Bridge         [32]byte
		RequestId      [32]byte
		Selector       []byte
		ReceiveSide    [32]byte
		OppositeBridge [32]byte
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
		ReceiveSide    [32]byte
		OppositeBridge [32]byte
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
