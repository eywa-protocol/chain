package test

import (
	"testing"

	"github.com/eywa-protocol/bls-crypto/bls"

	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/account"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/payload"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/signature"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
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
		Payload:    &payload.Epoch{Code: []byte("Chain Id")},
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
	t.Log(tx.TxType)
	err = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	t.Log(tx.TxType)
}
