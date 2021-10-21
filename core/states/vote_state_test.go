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
package states

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"testing"

	"bytes"

	"github.com/stretchr/testify/assert"
)

func TestVoteState_Deserialize_Serialize(t *testing.T) {
	_, pubKey1 := bls.GenerateRandomKey()
	_, pubKey2 := bls.GenerateRandomKey()

	vs := VoteState{
		StateBase:  StateBase{(byte)(1)},
		PublicKeys: []bls.PublicKey{pubKey1, pubKey2},
		Count:      10,
	}

	buf := bytes.NewBuffer(nil)
	vs.Serialize(buf)
	bs := buf.Bytes()

	var vs2 VoteState
	vs2.Deserialize(buf)
	assert.Equal(t, vs, vs2)

	buf = bytes.NewBuffer(bs[:len(bs)-1])
	err := vs2.Deserialize(buf)
	assert.NotNil(t, err)
}
