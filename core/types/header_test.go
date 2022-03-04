package types

import (
	"testing"

	"github.com/eywa-protocol/bls-crypto/bls"
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
		Signature:        bls.NewZeroMultisig(),
	}

	sink := common.NewZeroCopySink(nil)
	err := header.Serialization(sink)
	assert.NoError(t, err)
	bs := sink.Bytes()

	var h2 Header
	source := common.NewZeroCopySource(bs)
	err = h2.Deserialization(source)
	assert.Equal(t, header, h2)
	assert.NoError(t, err)
}
