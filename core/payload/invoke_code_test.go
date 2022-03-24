package payload

import (
	"testing"

	"github.com/eywa-protocol/chain/common"

	"github.com/stretchr/testify/assert"
)

func TestInvokeCode_Serialize(t *testing.T) {
	code := InvokeCode{
		Code: []byte{1, 2, 3},
	}
	sink := common.NewZeroCopySink(nil)
	err := code.Serialization(sink)
	assert.NoError(t, err)
	var code2 InvokeCode
	err = code2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, code, code2)
}
