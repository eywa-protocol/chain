package payload

import (
	"fmt"
	"github.com/eywa-protocol/chain/common"
)

// InvokeCode is an implementation of transaction payload for invoke smartcontract
type InvokeCode struct {
	Code []byte
}

//note: InvokeCode.Code has data reference of param source
func (self *InvokeCode) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}

	self.Code = code
	return nil
}

func (self *InvokeCode) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(self.Code)
}
