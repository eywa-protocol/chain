package payload

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
	"github.com/near/borsh-go"
)

type ReceiveRequestEvent struct {
	OriginData wrappers.BridgeReceiveRequest
}

func (self *ReceiveRequestEvent) Deserialization(source *common.ZeroCopySource) error {
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

func (self *ReceiveRequestEvent) Serialization(sink *common.ZeroCopySink) {
	oracleRequestBytes, _ := marshalBinaryRecievRequest(&self.OriginData)
	sink.WriteVarBytes(oracleRequestBytes)
}

func marshalBinaryRecievRequest(be *wrappers.BridgeReceiveRequest) (data []byte, err error) {
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
