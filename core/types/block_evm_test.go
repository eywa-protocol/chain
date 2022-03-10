package types

import (
	"crypto/ecdsa"
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
	"github.com/stretchr/testify/assert"
)

var (
	backend                        *backends.SimulatedBackend
	owner                          *bind.TransactOpts
	blockTest                      *wrappers.BlockTest
	err                            error
	ownerKey, signerKey            *ecdsa.PrivateKey
	ownerAddress, blockTestAddress ethcommon.Address
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

	blockTestAddress, _, blockTest, err = wrappers.DeployBlockTest(owner, backend)
	if err != nil {
		panic(err)
	}

	backend.Commit()
}

func Test_EvmMerkleProve(t *testing.T) {
	hash := common.Uint256{0xCA, 0xFE, 0xBA, 0xBE}

	header := Header{
		ChainID:          1111,
		PrevBlockHash:    hash,
		EpochBlockHash:   hash,
		TransactionsRoot: hash,
		SourceHeight:     100,
		Height:           10,
	}
	blockHash := header.Hash()

	solHash, err := blockTest.BlockHash(
		&bind.CallOpts{},
		header.ChainID,
		header.PrevBlockHash,
		header.EpochBlockHash,
		header.TransactionsRoot,
		header.SourceHeight,
		header.Height,
	)

	assert.NoError(t, err)
	assert.Equal(t, solHash[:], blockHash.ToArray())
}
