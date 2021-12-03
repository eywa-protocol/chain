package payload

import (
	"fmt"
	"github.com/eywa-protocol/chain/common"
)

type Epoch struct {
	Data []byte
}

//note: InvokeCode.Code has data reference of param source
func (self *Epoch) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}

	self.Data = code
	return nil
}

func (self *Epoch) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(self.Data)
}
