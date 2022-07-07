package genesis

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/account"
	"github.com/eywa-protocol/chain/cmd/utils"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/ledger"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/eywa-protocol/wrappers"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
)

// TODO: fix unhandled errors, remove unused acc if not needed

var (
	lg, lg2                     *ledger.Ledger
	err                         error
	acc                         *account.Account
	dbDir1, dbDir2, genesisFile string
	genesisBlock                *types.Block
)

func TestMain(m *testing.M) {
	dbDir1 = utils.GetStoreDirPath("test1", "")
	dbDir2 = utils.GetStoreDirPath("test2", "")
	genesisFile = "genesis.data"
	lg, _ = ledger.NewLedger(dbDir1, 1111)
	lg2, _ = ledger.NewLedger(dbDir2, 1112)
	acc = account.NewAccount(0)

	genesisBlock, err = BuildGenesisBlock(0, 0)
	if err != nil {
		fmt.Printf("BuildGenesisBlock error:%s\n", err)
	}
	err = saveBlockToFile(genesisBlock, genesisFile)
	if err != nil {
		fmt.Printf("saveBlockToFile error:%s\n", err)
	}
	err = lg.Init(genesisBlock)
	if err != nil {
		fmt.Printf("lg.Init error:%s\n", err)
	}
	m.Run()
	lg.Close()
	lg2.Close()
	os.RemoveAll(dbDir1)
	os.RemoveAll(dbDir2)
	os.RemoveAll(genesisFile)
}

func saveBlockToFile(block *types.Block, file string) (err error) {
	var f *os.File
	f, err = os.Create(file)
	defer f.Close()
	if err != nil {
		return err
	}
	genesisBlockBytes, _ := block.ToArray()
	err = ioutil.WriteFile(file, genesisBlockBytes, os.ModePerm)
	defer f.Close()
	if err != nil {
		return err
	}
	return nil
}

func TestLedgerInited(t *testing.T) {
	genBytes, err := ioutil.ReadFile(genesisFile)
	require.NoError(t, err)

	genesisBlockFromBytes, err := types.BlockFromRawBytes(genBytes)
	require.NoError(t, err)

	found, err := lg.IsContainBlock(genesisBlockFromBytes.Hash())
	require.NoError(t, err)
	require.True(t, found)

	block, err := lg.GetBlockByHash(genesisBlockFromBytes.Hash())
	require.Equal(t, genesisBlockFromBytes, block)
	require.Equal(t, genesisBlock, block)
}

func Test_BlockFromRawBytes(t *testing.T) {

	genBytes, err := ioutil.ReadFile(genesisFile)
	require.NoError(t, err)

	genesisBlockFromBytes, err := types.BlockFromRawBytes(genBytes)
	require.NoError(t, err)

	assert.Equal(t, genesisBlockFromBytes.Hash(), genesisBlock.Hash())
	assert.Equal(t, genesisBlockFromBytes, genesisBlock)
}

func Test_GetGenesisFromFileAndInitNewLedger(t *testing.T) {

	genBytes, err := ioutil.ReadFile(genesisFile)
	require.NoError(t, err)
	bFromBytes, err := types.BlockFromRawBytes(genBytes)
	require.NoError(t, err)
	assert.Equal(t, bFromBytes.Hash(), genesisBlock.Hash())
	err = lg2.Init(bFromBytes)
	require.NoError(t, err)
	require.Equal(t, lg.GetCurrentBlockHash(), lg2.GetCurrentBlockHash())
}

func TestGenesisBlockInit(t *testing.T) {
	block, err := BuildGenesisBlock(0, 0)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, block.Header.TransactionsRoot, common.UINT256_EMPTY)
	assert.Zero(t, len(block.Transactions))
}

func TestSaveBridgeEventAsBlock(t *testing.T) {
	blockBefore := lg.GetCurrentBlockHash()
	event := payload.NewBridgeEvent(&wrappers.BridgeOracleRequest{
		RequestType: "setRequest",
		Bridge:      ethCommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
		ChainId:     big.NewInt(94),
	})

	solEvent := payload.NewBridgeSolanaEvent(&wrappers.BridgeOracleRequestSolana{
		RequestType: "setRequest",
		Bridge:      [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 90, 1, 2, 3, 4, 5, 6, 7, 78, 9, 0, 1, 2, 2, 3, 43, 4, 4, 5, 5, 56, 23},
		ChainId:     big.NewInt(94),
	})

	sol2EVMEvent := payload.NewSolanaToEVMEvent(&bridge.BridgeEvent{
		OracleRequest: bridge.OracleRequest{
			RequestType:    "test",
			BridgePubKey:   solana.PublicKey{},
			RequestId:      solana.PublicKey{},
			Selector:       []byte("test"),
			ReceiveSide:    [20]uint8{},
			OppositeBridge: [20]uint8{},
			ChainId:        0,
		},
		Signature: solana.Signature{},
		Slot:      20,
	})

	solReceiveRequest := payload.NewSolReceiveRequestEvent(&bridge.BridgeReceiveEvent{
		ReceiveRequest: bridge.ReceiveRequest{
			RequestId:   solana.PublicKey{},
			ReceiveSide: solana.PublicKey{},
			BridgeFrom:  [20]uint8{},
		},
		Signature: solana.Signature{},
		Slot:      21,
	})

	epoch, err := bls.ReadPublicKey("1d65becbb891b6e69951febbc4ac066343670b34d84777a077c06871beb9c07f28be8a6fa825e9d615f56f0dbcd728b46e42b4ae2a611e2ab919a1de923ae7ed0f1c89b508af036f52c2215a04e13a7a5e891d9220d3d8751dc0525b81fca3051dc2e58a167c412941bd1adeb29f5a0beb5d26e748e8ca55e508deadead1ea5e")
	assert.NoError(t, err)
	epochEvent := payload.NewEpochEvent(123, common.UINT256_EMPTY, []bls.PublicKey{epoch, epoch, epoch}, []string{"one", "two", "three"})

	var txs types.Transactions
	txs = append(txs, types.ToTransaction(event))
	txs = append(txs, types.ToTransaction(solEvent))
	txs = append(txs, types.ToTransaction(sol2EVMEvent))
	txs = append(txs, types.ToTransaction(epochEvent))
	txs = append(txs, types.ToTransaction(solReceiveRequest))
	_, err = lg.CreateBlockFromEvents(txs, 123, common.UINT256_EMPTY)

	require.NoError(t, err)
	blockAfter := lg.GetCurrentBlockHash()
	require.Equal(t, blockBefore, blockAfter)

	t.Log("end")
}
