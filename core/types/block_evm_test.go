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
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/wrappers"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
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

	header.сalculateHash()
	blockHash := header.Hash()

	assert.NoError(t, err)
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
	assert.NotNil(t, solHash)
	assert.NotNil(t, blockHash)
	assert.Equal(t, solHash[:], blockHash.ToArray())
}

func Test_EvmHeaderRawData(t *testing.T) {
	hash := common.Uint256{0xCA, 0xFE, 0xBA, 0xBE}

	header := Header{
		ChainID:          1111,
		PrevBlockHash:    hash,
		EpochBlockHash:   hash,
		TransactionsRoot: hash,
		SourceHeight:     100,
		Height:           10,
	}

	header.сalculateHash()
	blockHash := header.Hash()
	assert.NoError(t, err)
	res, err := blockTest.BlockHeaderRawDataTest(&bind.CallOpts{}, header.RawData())
	assert.NoError(t, err)
	//t.Log(res)
	assert.Equal(t, res.AllBlockHash[:], blockHash.ToArray())
	assert.Equal(t, res.BlockTxHash[:], header.PrevBlockHash.ToArray())
}

func Test_EvmOracleRequestTxRawData(t *testing.T) {
	event := &payload.BridgeEvent{
		OriginData: wrappers.BridgeOracleRequest{
			RequestType: "setRequest",
			Bridge:      ethcommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
			ChainId:     big.NewInt(94),
			Selector:    []byte("my selector"),
		},
	}

	res, err := blockTest.OracleRequestTxRawDataTest(&bind.CallOpts{}, event.RawData())
	assert.NoError(t, err)

	tx := ToTransaction(event)
	hash := tx.Hash()
	assert.Equal(t, res.TxHash[:], hash.ToArray())
	assert.Equal(t, res.ReqId, event.OriginData.RequestId)
	assert.Equal(t, res.BridgeFrom[:20], event.OriginData.Bridge[:])
	assert.Equal(t, res.ReceiveSide, event.OriginData.ReceiveSide)
	assert.Equal(t, res.Sel, event.OriginData.Selector)
}

func Test_EvmSolanaRequestTxRawData(t *testing.T) {
	event := &payload.BridgeSolanaEvent{
		OriginData: wrappers.BridgeOracleRequestSolana{
			RequestType: "setRequest",
			Bridge:      [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 90, 1, 2, 3, 4, 5, 6, 7, 78, 9, 0, 1, 2, 2, 3, 43, 4, 4, 5, 5, 56, 23},
			ChainId:     big.NewInt(94),
			Selector:    []byte("my selector"),
		},
	}

	res, err := blockTest.SolanaRequestTxRawDataTest(&bind.CallOpts{}, event.RawData())
	assert.NoError(t, err)

	tx := ToTransaction(event)
	hash := tx.Hash()
	assert.Equal(t, res.TxHash[:], hash.ToArray())
	assert.Equal(t, res.ReqId, event.OriginData.RequestId)
	assert.Equal(t, res.BridgeFrom, event.OriginData.Bridge)
	assert.Equal(t, res.OppositeBridge, event.OriginData.OppositeBridge)
	assert.Equal(t, res.Sel, event.OriginData.Selector)
}

func Test_EvmSolanaToEvmRequestTxRawData(t *testing.T) {
	event := &payload.SolanaToEVMEvent{
		OriginData: bridge.BridgeEvent{
			OracleRequest: bridge.OracleRequest{
				RequestType:    "test",
				BridgePubKey:   solana.PublicKey{1, 2, 3},
				RequestId:      solana.PublicKey{10, 11, 12},
				Selector:       []byte("my selector"),
				ReceiveSide:    common.Address{20, 21, 22},
				OppositeBridge: common.Address{30, 31, 32},
				ChainId:        uint64(3),
			},
			Signature: solana.Signature{},
			Slot:      uint64(3),
		},
	}

	// BridgeEvent and SolanaToEvmEvent RawData must be binary compartible
	res, err := blockTest.OracleRequestTxRawDataTest(&bind.CallOpts{}, event.RawData())
	assert.NoError(t, err)

	tx := ToTransaction(event)
	hash := tx.Hash()
	assert.Equal(t, res.TxHash[:], hash.ToArray())
	assert.Equal(t, res.ReqId[:], event.OriginData.RequestId[:])
	assert.Equal(t, res.BridgeFrom[:], event.OriginData.BridgePubKey[:])
	assert.Equal(t, res.ReceiveSide[:], event.OriginData.ReceiveSide[:])
	assert.Equal(t, res.Sel, event.OriginData.Selector)
}

func Test_EvmEpochRequestTxRawData(t *testing.T) {
	epoch, err := bls.ReadPublicKey("1d65becbb891b6e69951febbc4ac066343670b34d84777a077c06871beb9c07f28be8a6fa825e9d615f56f0dbcd728b46e42b4ae2a611e2ab919a1de923ae7ed0f1c89b508af036f52c2215a04e13a7a5e891d9220d3d8751dc0525b81fca3051dc2e58a167c412941bd1adeb29f5a0beb5d26e748e8ca55e508deadead1ea5e")
	assert.NoError(t, err)

	event := payload.NewEpochEvent(123, common.UINT256_EMPTY, []bls.PublicKey{epoch, epoch, epoch}, []string{"one", "two", "three"})
	res, err := blockTest.EpochRequestTxRawDataTest(&bind.CallOpts{}, event.RawData())
	assert.NoError(t, err)

	tx := ToTransaction(event)
	hash := tx.Hash()
	assert.Equal(t, res.TxHash[:], hash.ToArray())
	assert.Equal(t, res.TxNewKey[:], event.EpochPublicKey.Marshal())
	assert.Equal(t, res.TxNewEpochParticipantsNum, uint8(len(event.PublicKeys)))
	assert.Equal(t, res.TxNewEpochNum, event.Number)
}

func TestEvmBlockMerkleProve(t *testing.T) {
	hash := common.Uint256{0xCA, 0xFE, 0xBA, 0xBE}

	payloads := []payload.BridgeEvent{
		{
			OriginData: wrappers.BridgeOracleRequest{
				RequestType: "setRequest",
				Bridge:      ethcommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
				RequestId:   payload.RequestId{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				Selector:    []byte{51, 52, 53},
				ReceiveSide: ethcommon.HexToAddress("0x2122232425262728293031323334353637383940"),
				ChainId:     big.NewInt(1111),
			},
		},
		{
			OriginData: wrappers.BridgeOracleRequest{
				RequestType: "setRequest",
				Bridge:      ethcommon.HexToAddress("0x0c760E9A85d3E957Dd1E189516b6658CfEcD3985"),
				RequestId:   payload.RequestId{1, 3, 3, 4, 5, 6, 7, 8, 9, 10, 1, 3, 3, 4, 5, 6, 7, 8, 9, 10, 1, 3, 3, 4, 5, 6, 7, 8, 9, 10, 11, 13},
				Selector:    []byte{51, 53, 53, 54, 55},
				ReceiveSide: ethcommon.HexToAddress("0x3133333435363738393031333334353637383940"),
				ChainId:     big.NewInt(1111),
			},
		},
		{
			OriginData: wrappers.BridgeOracleRequest{
				RequestType: "setRequest",
				Bridge:      ethcommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD4985"),
				RequestId:   payload.RequestId{1, 2, 4, 4, 5, 6, 7, 8, 9, 10, 1, 2, 4, 4, 5, 6, 7, 8, 9, 10, 1, 2, 4, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				Selector:    []byte{51},
				ReceiveSide: ethcommon.HexToAddress("0x2122242425262728294041424444454647484940"),
				ChainId:     big.NewInt(1111),
			},
		},
	}

	txs := Transactions{ToTransaction(&payloads[0]), ToTransaction(&payloads[1]), ToTransaction(&payloads[2])}
	block := NewBlock(1111, hash, hash, 100, 10, txs)

	for i := range block.Transactions {
		path, err := block.MerkleProve(i)
		assert.NoError(t, err)
		// t.Log(path)

		// Verify the merkle prove in evm smart contract
		res, err := merkleTest.BlockMerkleProveTest(&bind.CallOpts{}, path, block.Header.TransactionsRoot)
		assert.NoError(t, err)
		// t.Log(res)

		assert.Equal(t, res.BridgeFrom[:20], payloads[i].OriginData.Bridge[:])
		assert.Equal(t, res.ReqId, payloads[i].OriginData.RequestId)
		assert.Equal(t, res.Sel, payloads[i].OriginData.Selector)
		assert.Equal(t, res.ReceiveSide, payloads[i].OriginData.ReceiveSide)
	}
}
