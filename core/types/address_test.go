package types

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TODO TestAddressFromEpochValidators

func TestAddressFromEpochValidators(t *testing.T) {
	_, pubKey1 := bls.GenerateRandomKey()
	_, pubKey2 := bls.GenerateRandomKey()
	_, pubKey3 := bls.GenerateRandomKey()
	pubkeys := []bls.PublicKey{pubKey1, pubKey2, pubKey3}

	addr, _ := AddressFromPubLeySlice(pubkeys)
	addr2, _ := AddressFromMultiPubKeys(pubkeys, 3)
	assert.Equal(t, addr, addr2)

	pubkeys = []bls.PublicKey{pubKey3, pubKey2, pubKey1}

	addr3, _ := AddressFromMultiPubKeys(pubkeys, 3)

	assert.Equal(t, addr3, addr2)
}
