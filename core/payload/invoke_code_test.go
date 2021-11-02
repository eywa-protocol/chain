package payload

import (
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvokeCode_Serialize(t *testing.T) {
	code := InvokeCode{
		Code: []byte{1, 2, 3},
	}
	sink := common.NewZeroCopySink(nil)
	code.Serialization(sink)
	var code2 InvokeCode
	err := code2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, code, code2)
}
