/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/
package signature

import (
	"testing"

	"github.com/eywa-protocol/bls-crypto/bls"

	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/account"
)

func TestSign(t *testing.T) {

	acc := account.NewAccount(0)
	data := []byte{1, 2, 3}

	sig, err := Sign(acc, data)
	assert.Nil(t, err)

	sig2, err := bls.UnmarshalSignature(sig)
	assert.Nil(t, err)

	verified := sig2.Verify(acc.PublicKey, data)
	assert.True(t, verified)

	err = Verify(acc.PublicKey, data, sig2)
	assert.Nil(t, err)

}

func TestSignature(t *testing.T) {
	acc := account.NewAccount(0)
	data := []byte{1, 2, 3}
	sig, err := Signature(acc, data)
	assert.Nil(t, err)

	verified := sig.Verify(acc.PublicKey, data)
	assert.True(t, verified)

}

// Multisignature is verified in TestMultiVerifyTx
