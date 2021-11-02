package test

import (
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"testing"
)

func TestFixed64_Serialize(t *testing.T) {
	val := common.Fixed64(10)
	buf := common.NewZeroCopySink(nil)
	val.Serialization(buf)
	val2 := common.Fixed64(0)
	val2.Deserialization(common.NewZeroCopySource(buf.Bytes()))

	assert.Equal(t, val, val2)
}

func TestFixed64_Deserialize(t *testing.T) {
	val := common.Fixed64(0)
	err := val.Deserialization(common.NewZeroCopySource([]byte{1, 2, 3}))

	assert.NotNil(t, err)

}
