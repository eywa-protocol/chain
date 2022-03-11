package payload

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
)

type BridgeSolanaEvent struct {
	OriginData wrappers.BridgeOracleRequestSolana
}

func (tx *BridgeSolanaEvent) TxType() TransactionType {
	return BridgeEventSolanaType
}

func (self *BridgeSolanaEvent) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}
	err := unmarshalBinarySolana(code, &self.OriginData)
	if err != nil {
		return err
	}
	return nil
}

func (self *BridgeSolanaEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := MarshalSolBinary(&self.OriginData)
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
		Chainid        *big.Int
		Raw            types.Log
	}
	gob.NewDecoder(r).Decode(&dec)
	st.RequestType = dec.RequestType

	st.Bridge = dec.Bridge
	st.RequestId = dec.RequestId
	st.Selector = dec.Selector
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
		OppositeBridge [32]byte
		Chainid        *big.Int
		Raw            types.Log
	}{
		be.RequestType,
		be.Bridge,
		be.RequestId,
		be.Selector,
		be.OppositeBridge,
		be.Chainid,
		be.Raw,
	}); err != nil {
		return nil, err
	}

	w.Flush()
	return b.Bytes(), nil
}

func (self *BridgeSolanaEvent) Hash() common.Uint256 {
	var data []byte
	data = append(data, self.OriginData.Bridge[12:]...) // TODO [:]
	data = append(data, self.OriginData.RequestId[:]...)
	data = append(data, self.OriginData.Selector...)
	data = append(data, self.OriginData.OppositeBridge[:]...)
	return sha256.Sum256(data)
}
