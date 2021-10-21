package test

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/account"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/cmd/utils"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/payload"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/signature"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
	"testing"
)

func TestSign(t *testing.T) {
	acc := account.NewAccount("")
	data := []byte{1, 2, 3}

	sig, err := utils.Sign(data, acc)
	assert.Nil(t, err)

	err = signature.Verify(acc.PublicKey, data, sig)
	assert.Nil(t, err)
}

func testVerifyTx(t *testing.T) {
	acc1 := account.NewAccount("123")

	tx := &types.Transaction{
		Version:    0,
		TxType:     types.TransactionType(types.Invoke),
		Nonce:      1,
		ChainID:    0,
		Payload:    &payload.InvokeCode{Code: []byte("Chain Id")},
		Attributes: []byte{},
	}
	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	assert.NoError(t, err)

	tx, err = types.TransactionFromRawBytes(sink.Bytes())
	assert.NoError(t, err)

	err = utils.SignTransaction(acc1, tx)
	assert.NoError(t, err)

	hash := tx.Hash()
	err = signature.Verify(acc1.PublicKey, hash.ToArray(), tx.Sigs[0].SigData[0])
	assert.NoError(t, err)

	addr, err := tx.GetSignatureAddresses()
	assert.NoError(t, err)
	assert.Equal(t, acc1.Address, addr[0])
}

func testMultiVerifyTx(t *testing.T) {
	acc1 := account.NewAccount("")
	acc2 := account.NewAccount("")
	acc3 := account.NewAccount("")

	accAddr, err := types.AddressFromMultiPubKeys([]bls.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey}, 2)
	assert.NoError(t, err)
	tx := &types.Transaction{
		Version:    0,
		TxType:     types.TransactionType(types.Invoke),
		Nonce:      1,
		ChainID:    0,
		Payload:    &payload.InvokeCode{Code: []byte("Chain Id")},
		Attributes: []byte{},
	}
	sink := common.NewZeroCopySink(nil)
	err = tx.Serialization(sink)
	assert.NoError(t, err)

	tx, err = types.TransactionFromRawBytes(sink.Bytes())
	assert.NoError(t, err)

	err = utils.MultiSigTransaction(tx, 2, []bls.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey}, acc1)
	assert.NoError(t, err)

	err = utils.MultiSigTransaction(tx, 2, []bls.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey}, acc2)
	assert.NoError(t, err)

	hash := tx.Hash()
	err = signature.VerifyMultiSignature(hash.ToArray(), tx.Sigs[0].PubKeys, 2, tx.Sigs[0].SigData)
	assert.NoError(t, err)

	addr, err := tx.GetSignatureAddresses()
	assert.NoError(t, err)
	assert.Equal(t, accAddr, addr[0])
}
