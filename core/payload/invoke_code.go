package payload

import (
	"encoding/json"
	"fmt"

	"github.com/eywa-protocol/chain/common"
)

// InvokeCode DEPRECATED not used by EYWA bridge and will be removed in future
// todo: remove with dependencies
type InvokeCode struct {
	Code []byte
}

func (e *InvokeCode) TxType() TransactionType {
	return InvokeType
}

func (e *InvokeCode) RequestState() RequestState {
	return ReqStateUnknown
}

func (e *InvokeCode) RequestId() RequestId {
	return RequestId{}
}

func (e *InvokeCode) ToJson() (json.RawMessage, error) {
	// TODO implement me
	panic("implement me")
}

func (e *InvokeCode) SrcTxHash() []byte {
	// TODO implement me
	panic("implement me")
}

func (e *InvokeCode) DstChainId() (uint64, bool) {
	// TODO implement me
	panic("implement me")
}

func (e *InvokeCode) Deserialization(source *common.ZeroCopySource) error {
	code, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("[InvokeCode] deserialize code error")
	}

	e.Code = code
	return nil
}

func (e *InvokeCode) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes(e.Code)
	return nil
}

func (e *InvokeCode) RawData() []byte {
	return e.Code
}
