package types

import (
	"fmt"
	"testing"

	"github.com/eywa-protocol/chain/common"
	"github.com/stretchr/testify/assert"
)

func TestHeader_Serialize(t *testing.T) {
	header := Header{
		ChainID:          123,
		PrevBlockHash:    common.UINT256_EMPTY,
		EpochBlockHash:   common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		SourceHeight:     123,
		Height:           123,
	}

	sink := common.NewZeroCopySink(nil)
	header.Serialization(sink)
	bs := sink.Bytes()

	var h2 Header
	source := common.NewZeroCopySource(bs)
	err := h2.Deserialization(source)
	assert.Equal(t, fmt.Sprint(header), fmt.Sprint(h2))

	assert.Nil(t, err)
}
