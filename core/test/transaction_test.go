package test

import (
	"testing"

	"github.com/eywa-protocol/bls-crypto/bls"

	"github.com/eywa-protocol/chain/account"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/signature"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Serialize(t *testing.T) {
	tx := &types.Transaction{
		TxType:     types.TransactionType(types.Invoke),
		Nonce:      1,
		ChainID:    0,
		Payload:    &payload.InvokeCode{Code: []byte("Chain Id")},
		Attributes: []byte("Chain Id"),
	}

	sink := common.NewZeroCopySink(nil)
	err := tx.SerializeUnsigned(sink)
	assert.Error(t, err)

	tx.Attributes = []byte{}

	err = tx.SerializeUnsigned(sink)
	assert.NoError(t, err)

	acc := account.NewAccount(0)
	sigData, err := signature.Sign(acc, sink.Bytes())
	assert.NoError(t, err)

	sig, _ := bls.UnmarshalSignature(sigData)
	tx.Sig = types.Sig{
		SigData: sig,
		M:       1,
	}
	sink.Reset()

	err = tx.Serialization(sink)
	assert.NoError(t, err)

	tx = new(types.Transaction)
	err = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
}

func TestEpochTransaction_Serialize(t *testing.T) {
	tx := &types.Transaction{
		TxType:     types.Epoch,
		Nonce:      1,
		ChainID:    0,
		Payload:    &payload.Epoch{Data: []byte("Chain Id")},
		Attributes: []byte("Chain Id"),
	}

	sink := common.NewZeroCopySink(nil)
	err := tx.SerializeUnsigned(sink)
	assert.Error(t, err)

	tx.Attributes = []byte{}

	err = tx.SerializeUnsigned(sink)
	assert.NoError(t, err)

	acc := account.NewAccount(0)
	sigData, err := signature.Sign(acc, sink.Bytes())
	assert.NoError(t, err)

	sig, _ := bls.UnmarshalSignature(sigData)
	tx.Sig = types.Sig{
		SigData: sig,
		M:       1,
	}
	sink.Reset()

	err = tx.Serialization(sink)
	assert.NoError(t, err)

	tx = new(types.Transaction)
	err = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	t.Log(tx.TxType)
}
