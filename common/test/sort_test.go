/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/
package test

import (
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"testing"
)

func TestUint64Slice(t *testing.T) {
	data1 := []uint64{3, 2, 4, 1}
	data2 := []uint64{1, 2, 3, 4}

	common.SortUint64s(data1)

	assert.Equal(t, data1, data2)

}
