/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
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
