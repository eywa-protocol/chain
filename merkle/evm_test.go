package merkle

import (
	"crypto/ecdsa"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/wrappers"
)

var (
	backend                         *backends.SimulatedBackend
	owner                           *bind.TransactOpts
	merkleTest                      *wrappers.MerkleTest
	err                             error
	ownerKey, signerKey             *ecdsa.PrivateKey
	ownerAddress, merkleTestAddress common.Address
)

func init() {
	ownerKey, _ = crypto.GenerateKey()

	signerKey, _ = crypto.GenerateKey()

	ownerAddress = crypto.PubkeyToAddress(ownerKey.PublicKey)

	genesis := core.GenesisAlloc{
		ownerAddress: {Balance: new(big.Int).SetInt64(math.MaxInt64)},
	}
	backend = backends.NewSimulatedBackend(genesis, math.MaxInt64)

	owner, err = bind.NewKeyedTransactorWithChainID(ownerKey, big.NewInt(1337))
	if err != nil {
		panic(err)
	}

	merkleTestAddress, _, merkleTest, err = wrappers.DeployMerkleTest(owner, backend)
	if err != nil {
		panic(err)
	}

	backend.Commit()
}

func TestEvmMerkleProve(t *testing.T) {
	n := uint32(10)
	store, _ := NewFileHashStore("merkletree.db", 0)
	defer func() { os.Remove("merkletree.db") }()
	tree := NewTree(0, nil, store)
	for i := uint32(0); i < n; i++ {
		tree.Append([]byte{byte(i + 1)})
	}
	root := tree.Root()

	for i := uint32(0); i < n; i++ {
		data := []byte{byte(i + 1)}
		path, err := tree.MerkleInclusionLeafPath(data, i, n)
		require.NoError(t, err)

		// Verify the merkle prove in go
		val, err := MerkleProve(path, root.ToArray())
		assert.Nil(t, err)
		assert.Equal(t, data, val)

		// Verify the merkle prove in evm smart contract
		val, err = merkleTest.MerkleProve(&bind.CallOpts{}, path, root)
		require.NoError(t, err)
		assert.Equal(t, data, val)
	}
}
