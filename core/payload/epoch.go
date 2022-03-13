package payload

import (
	"fmt"

	"github.com/eywa-protocol/chain/common"
)

type Epoch struct {
	Data []byte
}

func (tx *Epoch) TxType() TransactionType {
	return EpochType
}

func (self *Epoch) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}

	self.Data = code
	return nil
}

func (self *Epoch) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes(self.Data)
	return nil
}

func (self *Epoch) RawData() []byte {
	return self.Data
}
