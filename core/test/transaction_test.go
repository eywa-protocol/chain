package test

import (
	"testing"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Serialize(t *testing.T) {
	tx := &types.Transaction{
		TxType:  types.TransactionType(types.Invoke),
		Payload: &payload.InvokeCode{Code: []byte("Chain Id")},
	}

	sink := common.NewZeroCopySink(nil)
	err := tx.SerializeUnsigned(sink)
	assert.NoError(t, err)

	err = tx.SerializeUnsigned(sink)
	assert.NoError(t, err)

	sink.Reset()

	err = tx.Serialization(sink)
	assert.NoError(t, err)

	tx = new(types.Transaction)
	err = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
}

func TestEpochTransaction_Serialize(t *testing.T) {
	tx := &types.Transaction{
		TxType:  types.Epoch,
		Payload: &payload.Epoch{Data: []byte("Chain Id")},
	}

	sink := common.NewZeroCopySink(nil)
	err := tx.SerializeUnsigned(sink)
	assert.NoError(t, err)

	err = tx.SerializeUnsigned(sink)
	assert.NoError(t, err)

	sink.Reset()

	err = tx.Serialization(sink)
	assert.NoError(t, err)

	tx = new(types.Transaction)
	err = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	t.Log(tx.TxType)
}
