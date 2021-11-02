/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/

package test

import (
	"bytes"
	"crypto/rand"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressFromBase58(t *testing.T) {
	var addr common.Address
	rand.Read(addr[:])

	base58 := addr.ToBase58()
	b1 := string(append([]byte{'X'}, []byte(base58)...))
	_, err := common.AddressFromBase58(b1)

	assert.NotNil(t, err)

	b2 := string([]byte(base58)[1:10])
	_, err = common.AddressFromBase58(b2)

	assert.NotNil(t, err)
}

func TestAddressParseFromBytes(t *testing.T) {
	var addr common.Address
	rand.Read(addr[:])

	addr2, _ := common.AddressParseFromBytes(addr[:])

	assert.Equal(t, addr, addr2)
}

func TestAddress_Serialize(t *testing.T) {
	var addr common.Address
	rand.Read(addr[:])

	buf := bytes.NewBuffer(nil)
	addr.Serialize(buf)

	var addr2 common.Address
	addr2.Deserialize(buf)
	assert.Equal(t, addr, addr2)
}
