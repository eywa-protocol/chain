package test

import (
	"testing"

	"github.com/eywa-protocol/chain/common"
	"github.com/stretchr/testify/assert"
)

// TODO: fix unhandled errors

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
