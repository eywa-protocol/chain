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

func TestBookkeeper_Deserialize_Serialize(t *testing.T) {
	_, pubKey1 := bls.GenerateRandomKey()
	_, pubKey2 := bls.GenerateRandomKey()
	_, pubKey3 := bls.GenerateRandomKey()
	_, pubKey4 := bls.GenerateRandomKey()

	bk := BookkeeperState{
		StateBase:      StateBase{(byte)(1)},
		CurrBookkeeper: []bls.PublicKey{pubKey1, pubKey2},
		NextBookkeeper: []bls.PublicKey{pubKey3, pubKey4},
	}

	buf := bytes.NewBuffer(nil)
	bk.Serialize(buf)
	bs := buf.Bytes()

	var bk2 BookkeeperState
	bk2.Deserialize(buf)
	assert.Equal(t, bk, bk2)

	buf = bytes.NewBuffer(bs[:len(bs)-1])
	err := bk2.Deserialize(buf)
	assert.NotNil(t, err)
}
