package test

import (
	"testing"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/account"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/cmd/utils"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/payload"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/signature"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
)

func TestSign(t *testing.T) {
	acc := account.NewAccount(0)
	data := []byte{1, 2, 3}

	sig, err := utils.Sign(data, acc)
	assert.Nil(t, err)

	err = signature.Verify(acc.PublicKey, data, sig)
	assert.Nil(t, err)
}

func TestVerifyTx(t *testing.T) {
	acc1 := account.NewAccount(0)

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
	err = signature.Verify(acc1.PublicKey, hash.ToArray(), tx.Sig.SigData)
	assert.NoError(t, err)

	// addr, err := tx.GetSignatureAddresses()
	// assert.NoError(t, err)
	// assert.Equal(t, acc1.Address, addr[0])
}

func TestMultiVerifyTx(t *testing.T) {
	acc1 := account.NewAccount(0)
	acc2 := account.NewAccount(1)
	acc3 := account.NewAccount(2)

	// Anti-rogue key attack coefficients
	as := bls.CalculateAntiRogueCoefficients([]bls.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey})

	// Aggregated public key of all participants
	allPub := bls.AggregatePublicKeys([]bls.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey}, as)

	// Setup phase - generate membership keys
	mk1 := acc1.PrivateKey.GenerateMembershipKeyPart(0, allPub, as[0]).
		Aggregate(acc2.PrivateKey.GenerateMembershipKeyPart(0, allPub, as[1])).
		Aggregate(acc3.PrivateKey.GenerateMembershipKeyPart(0, allPub, as[2]))
	mk2 := acc1.PrivateKey.GenerateMembershipKeyPart(1, allPub, as[0]).
		Aggregate(acc2.PrivateKey.GenerateMembershipKeyPart(1, allPub, as[1])).
		Aggregate(acc3.PrivateKey.GenerateMembershipKeyPart(1, allPub, as[2]))
	// mk3 := acc1.PrivateKey.GenerateMembershipKeyPart(2, allPub, Simple).
	// 	Aggregate(acc2.PrivateKey.GenerateMembershipKeyPart(2, allPub, Simple)).
	// 	Aggregate(acc3.PrivateKey.GenerateMembershipKeyPart(2, allPub, Simple))

	//accAddr, err := types.AddressFromMultiPubKeys([]bls.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey}, 2)
	//assert.NoError(t, err)
	tx := &types.Transaction{
		Version:    0,
		TxType:     types.TransactionType(types.Invoke),
		Nonce:      1,
		ChainID:    0,
		Payload:    &payload.InvokeCode{Code: []byte("Chain Id")},
		Attributes: []byte{},
		Sig:        types.Sig{bls.ZeroSignature(), bls.ZeroPublicKey(), 0}, // FIXME
	}
	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	assert.NoError(t, err)

	tx, err = types.TransactionFromRawBytes(sink.Bytes())
	assert.NoError(t, err)

	err = utils.MultiSigTransaction(tx, mk1, allPub, acc1)
	assert.NoError(t, err)

	err = utils.MultiSigTransaction(tx, mk2, allPub, acc2)
	assert.NoError(t, err)

	hash := tx.Hash()
	err = signature.VerifyMultiSignature(hash.ToArray(), tx.Sig.SigData, allPub, tx.Sig.PubKey, int64(tx.Sig.M))
	assert.NoError(t, err)

	//addr, err := tx.GetSignatureAddresses()
	//assert.NoError(t, err)
	//assert.Equal(t, accAddr, addr[0])
}
