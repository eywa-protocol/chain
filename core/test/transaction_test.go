package test

import (
	"testing"

	"github.com/eywa-protocol/bls-crypto/bls"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Serialize(t *testing.T) {
	tx := types.ToTransaction(&payload.InvokeCode{Code: []byte("Chain Id")})

	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	assert.NoError(t, err)

	sink.Reset()

	err = tx.Serialization(sink)
	assert.NoError(t, err)

	tx1, err := types.TransactionDeserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, tx, tx1)
}

func TestEpochTransaction_Serialize(t *testing.T) {
	epoch, err := bls.ReadPublicKey("1d65becbb891b6e69951febbc4ac066343670b34d84777a077c06871beb9c07f28be8a6fa825e9d615f56f0dbcd728b46e42b4ae2a611e2ab919a1de923ae7ed0f1c89b508af036f52c2215a04e13a7a5e891d9220d3d8751dc0525b81fca3051dc2e58a167c412941bd1adeb29f5a0beb5d26e748e8ca55e508deadead1ea5e")
	assert.NoError(t, err)

	tx := types.ToTransaction(&payload.Epoch{
		Number:         1,
		EpochPublicKey: epoch,
		PublicKeys:     []bls.PublicKey{},
	})

	sink := common.NewZeroCopySink(nil)
	err = tx.Serialization(sink)
	assert.NoError(t, err)

	err = tx.Serialization(sink)
	assert.NoError(t, err)

	sink.Reset()

	err = tx.Serialization(sink)
	assert.NoError(t, err)

	tx1, err := types.TransactionDeserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, tx, tx1)
	// t.Log(tx.TxType())
}
