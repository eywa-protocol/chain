/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/
package states

import (
	"testing"

	"bytes"
	"crypto/rand"

	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
)

func TestStorageKey_Deserialize_Serialize(t *testing.T) {
	var addr common.Address
	rand.Read(addr[:])

	storage := StorageKey{
		ContractAddress: addr,
		Key:             []byte{1, 2, 3},
	}

	buf := bytes.NewBuffer(nil)
	storage.Serialize(buf)
	bs := buf.Bytes()

	var storage2 StorageKey
	storage2.Deserialize(buf)
	assert.Equal(t, storage, storage2)

	buf = bytes.NewBuffer(bs[:len(bs)-1])
	err := storage2.Deserialize(buf)
	assert.NotNil(t, err)
}
