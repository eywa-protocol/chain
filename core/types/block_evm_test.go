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
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/merkle"
	"github.com/eywa-protocol/wrappers"
	"github.com/stretchr/testify/assert"
)

var (
	backend             *backends.SimulatedBackend
	owner               *bind.TransactOpts
	blockTest           *wrappers.BlockTest
	merkleTest          *wrappers.MerkleTest
	err                 error
	ownerKey, signerKey *ecdsa.PrivateKey
	ownerAddress        ethcommon.Address
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

	_, _, blockTest, err = wrappers.DeployBlockTest(owner, backend)
	if err != nil {
		panic(err)
	}

	_, _, merkleTest, err = wrappers.DeployMerkleTest(owner, backend)
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

	res, err := blockTest.BlockHeaderRawDataTest(&bind.CallOpts{}, header.RawData())
	assert.NoError(t, err)
	// t.Log(res)
	assert.Equal(t, res.AllBlockHash[:], blockHash.ToArray())
	assert.Equal(t, res.BlockTxHash[:], header.PrevBlockHash.ToArray())
}

func TestEvmBlockMerkleProve(t *testing.T) {
	hash := common.Uint256{0xCA, 0xFE, 0xBA, 0xBE}

	payloads := []payload.BridgeEvent{
		{
			OriginData: wrappers.BridgeOracleRequest{
				RequestType: "setRequest",
				Bridge:      ethcommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
				RequestId:   [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				Selector:    []byte{51, 52, 53},
				ReceiveSide: ethcommon.HexToAddress("0x2122232425262728293031323334353637383940"),
				Chainid:     big.NewInt(1111),
			},
		},
		{
			OriginData: wrappers.BridgeOracleRequest{
				RequestType: "setRequest",
				Bridge:      ethcommon.HexToAddress("0x0c760E9A85d3E957Dd1E189516b6658CfEcD3985"),
				RequestId:   [32]byte{1, 3, 3, 4, 5, 6, 7, 8, 9, 10, 1, 3, 3, 4, 5, 6, 7, 8, 9, 10, 1, 3, 3, 4, 5, 6, 7, 8, 9, 10, 11, 13},
				Selector:    []byte{51, 53, 53, 54, 55},
				ReceiveSide: ethcommon.HexToAddress("0x3133333435363738393031333334353637383940"),
				Chainid:     big.NewInt(1111),
			},
		},
		{
			OriginData: wrappers.BridgeOracleRequest{
				RequestType: "setRequest",
				Bridge:      ethcommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD4985"),
				RequestId:   [32]byte{1, 2, 4, 4, 5, 6, 7, 8, 9, 10, 1, 2, 4, 4, 5, 6, 7, 8, 9, 10, 1, 2, 4, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				Selector:    []byte{51},
				ReceiveSide: ethcommon.HexToAddress("0x2122242425262728294041424444454647484940"),
				Chainid:     big.NewInt(1111),
			},
		},
	}

	txs := Transactions{ToTransaction(&payloads[0]), ToTransaction(&payloads[1]), ToTransaction(&payloads[2])}
	block := NewBlock(1111, hash, hash, 100, 10, txs)

	store := merkle.NewMemHashStore()
	tree := merkle.NewTree(0, nil, store)
	for _, tx := range block.Transactions {
		tree.Append(tx.Payload.RawData())
	}
	root := tree.Root()

	for i := range payloads {
		data := block.Transactions[i].Payload.RawData()
		// t.Log(len(data), data)

		path, err := tree.MerkleInclusionLeafPath(data, uint64(i), uint64(len(block.Transactions)))
		assert.NoError(t, err)
		// t.Log(path)

		// Verify the merkle prove in evm smart contract
		bridge, reqId, sel, receiveSide, err := merkleTest.BlockMerkleProveTest(&bind.CallOpts{}, path, root)
		assert.NoError(t, err)
		// t.Log(bridge, reqId, sel, receiveSide)

		assert.Equal(t, bridge, payloads[i].OriginData.Bridge)
		assert.Equal(t, reqId, payloads[i].OriginData.RequestId)
		assert.Equal(t, sel, payloads[i].OriginData.Selector)
		assert.Equal(t, receiveSide, payloads[i].OriginData.ReceiveSide)
	}
}
