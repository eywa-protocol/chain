package payload

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
	"github.com/near/borsh-go"
)

type BridgeEvent struct {
	OriginData wrappers.BridgeOracleRequest
}

func (tx *BridgeEvent) TxType() TransactionType {
	return BridgeEventType
}

func (self *BridgeEvent) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}
	err := borsh.Deserialize(&self.OriginData, code)
	if err != nil {
		return err
	}
	return nil
}

func (self *BridgeEvent) Serialization(sink *common.ZeroCopySink) error {
	oracleRequestBytes, err := MarshalBinary(&self.OriginData)
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

	w.Flush()
	return b.Bytes(), nil
}

func (self *BridgeEvent) RawData() []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteBytes(self.OriginData.Bridge[:])
	sink.WriteBytes(self.OriginData.RequestId[:])
	sink.WriteVarBytes(self.OriginData.Selector)
	sink.WriteBytes(self.OriginData.ReceiveSide[:])
	sink.WriteUint64(self.OriginData.Chainid.Uint64())
	return sink.Bytes()
}
