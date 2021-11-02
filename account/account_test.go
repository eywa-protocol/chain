package account

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
)

func TestNewAccount(t *testing.T) {
	defer func() {
		os.RemoveAll("Log/")
	}()

	names := []string{
		"",
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
		"10",
		"11",
	}
	accounts := make([]*Account, len(names))
	for k, _ := range names {
		accounts[k] = NewAccount(byte(k))
		assert.NotNil(t, accounts[k])
		assert.NotNil(t, accounts[k].PrivateKey)
		assert.NotNil(t, accounts[k].PublicKey)
		assert.NotNil(t, accounts[k].Address)
		assert.NotNil(t, accounts[k].PrivKey())
		assert.NotNil(t, accounts[k].PubKey())
		assert.Equal(t, accounts[k].Address, types.AddressFromPubKey(accounts[k].PublicKey))
	}
}
