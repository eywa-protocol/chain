package test

import (
	"math/big"
	"testing"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
	proof := types.CryptoProof{
		PartSignature: bls.ZeroSignature(),
		PartPublicKey: bls.ZeroPublicKey(),
		SigMask:       *big.NewInt(0),
	}
	header := types.Header{
		ChainID:          0,
		PrevBlockHash:    common.UINT256_EMPTY,
		EpochBlockHash:   common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		SourceHeight:     12,
		Height:           12,
		Signature:        proof,
	}

	sink := common.NewZeroCopySink(nil)
	err := header.Serialization(sink)
	assert.NoError(t, err)

	var h types.Header
	source := common.NewZeroCopySource(sink.Bytes())
	err = h.Deserialization(source)
	assert.NoError(t, err)

	assert.Equal(t, header, h)
}

func BenchmarkT1(b *testing.B) {
	proof := types.CryptoProof{
		PartSignature: bls.ZeroSignature(),
		PartPublicKey: bls.ZeroPublicKey(),
		SigMask:       *big.NewInt(0),
	}
	header := types.Header{
		ChainID:          0,
		PrevBlockHash:    common.UINT256_EMPTY,
		EpochBlockHash:   common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		SourceHeight:     12,
		Height:           12,
		Signature:        proof,
	}

	buf := common.NewZeroCopySink([]byte(""))
	header.Serialization(buf)
	for i := 0; i < b.N; i++ {
		var h types.Header
		err := h.Deserialization(common.NewZeroCopySource(buf.Bytes()))
		assert.NoError(b, err)
	}
}

func BenchmarkT3(b *testing.B) {
	proof := types.CryptoProof{
		PartSignature: bls.ZeroSignature(),
		PartPublicKey: bls.ZeroPublicKey(),
		SigMask:       *big.NewInt(0),
	}
	header := types.Header{
		ChainID:          0,
		PrevBlockHash:    common.UINT256_EMPTY,
		EpochBlockHash:   common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		SourceHeight:     12,
		Height:           12,
		Signature:        proof,
	}

	for i := 0; i < b.N; i++ {
		buf := common.NewZeroCopySink(nil)
		header.Serialization(buf)
		var h types.Header
		h.Deserialization(common.NewZeroCopySource(buf.Bytes()))
	}
}
