package test

import (
	"github.com/eywa-protocol/chain/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUint64Slice(t *testing.T) {
	data1 := []uint64{3, 2, 4, 1}
	data2 := []uint64{1, 2, 3, 4}

	common.SortUint64s(data1)

	assert.Equal(t, data1, data2)

}
