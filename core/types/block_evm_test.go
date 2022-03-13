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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
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

func Test_EvmHeaderHash(t *testing.T) {
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

func Test_EvmTxBridgeEventHash(t *testing.T) {
	data := wrappers.BridgeOracleRequest{
		Bridge:      ethcommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
		RequestType: "setRequest",
		RequestId:   [32]byte{0xDE, 0xAD, 0xBE, 0xEF},
		Selector:    []byte{1, 2, 3, 4, 5},
		ReceiveSide: ethcommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
		Chainid:     big.NewInt(94),
		Raw: types.Log{
			Topics: []ethcommon.Hash{},
			Data:   []uint8{},
		},
	}
	payload := payload.BridgeEvent{OriginData: data}
	tx := ToTransaction(&payload)
	txHash := tx.Hash()

	solHash, err := blockTest.OracleRequestTest(
		&bind.CallOpts{},
		payload.OriginData.Bridge,
		payload.OriginData.RequestId,
		payload.OriginData.Selector,
		payload.OriginData.ReceiveSide,
	)

	assert.NoError(t, err)
	assert.Equal(t, solHash[:], txHash.ToArray())
}

func Test_EvmTxBridgeEventSolanaHash(t *testing.T) {
	data := wrappers.BridgeOracleRequestSolana{
		Bridge:         [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 90, 1, 2, 3, 4, 5, 6, 7, 78, 9, 0, 1, 2, 2, 3, 43, 4, 4, 5, 5, 56, 23},
		RequestType:    "setRequest",
		RequestId:      [32]byte{0xDE, 0xAD, 0xBE, 0xEF},
		Selector:       []byte{1, 2, 3, 4, 5},
		OppositeBridge: [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 90, 1, 2, 3, 4, 5, 6, 7, 78, 9, 0, 1, 2, 2, 3, 43, 4, 4, 5, 5, 56, 23},
		Chainid:        big.NewInt(94),
		Raw: types.Log{
			Topics: []ethcommon.Hash{},
			Data:   []uint8{},
		},
	}
	payload := payload.BridgeSolanaEvent{OriginData: data}
	tx := ToTransaction(&payload)
	txHash := tx.Hash()

	solHash, err := blockTest.OracleRequestTestSolana(
		&bind.CallOpts{},
		payload.OriginData.Bridge,
		payload.OriginData.RequestId,
		payload.OriginData.Selector,
		payload.OriginData.OppositeBridge,
	)

	assert.NoError(t, err)
	assert.Equal(t, solHash[:], txHash.ToArray())
}

func Test_EvmHeaderRawDataHash(t *testing.T) {
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

	res, err := blockTest.RawDataTest(&bind.CallOpts{}, header.RawData())
	assert.NoError(t, err)
	assert.Equal(t, res.AllBlockHash[:], blockHash.ToArray())
	assert.Equal(t, res.BlockTxHash[:], header.PrevBlockHash.ToArray())
}
