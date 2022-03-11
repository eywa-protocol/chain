package test

import (
	"testing"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Serialize(t *testing.T) {
	var tx types.Transaction
	tx = &payload.InvokeCode{Code: []byte("Chain Id")}

	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	assert.NoError(t, err)

	sink.Reset()

	err = tx.Serialization(sink)
	assert.NoError(t, err)

	var tx1 payload.InvokeCode
	err = tx1.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, tx, &tx1)
}

func TestEpochTransaction_Serialize(t *testing.T) {
	var tx types.Transaction
	tx = &payload.Epoch{Data: []byte("Chain Id")}

	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	assert.NoError(t, err)

	err = tx.Serialization(sink)
	assert.NoError(t, err)

	sink.Reset()

	err = tx.Serialization(sink)
	assert.NoError(t, err)

	var tx1 payload.Epoch
	err = tx1.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, tx, &tx1)
	// t.Log(tx.TxType())
}
