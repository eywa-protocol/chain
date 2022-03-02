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

var (
	lg, lg2                     *ledger.Ledger
	err                         error
	acc                         *account.Account
	dbDir1, dbDir2, genesisFile string
	genesisBlock                *types.Block
	bookKeepers                 []bls.PublicKey
	bookKeepersBytes            []byte
)

func TestMain(m *testing.M) {
	dbDir1 = utils.GetStoreDirPath("test1", "")
	dbDir2 = utils.GetStoreDirPath("test2", "")
	genesisFile = "genesis.data"
	lg, _ = ledger.NewLedger(dbDir1)
	lg2, _ = ledger.NewLedger(dbDir2)
	acc = account.NewAccount(0)

	bookKeepers = []bls.PublicKey{acc.PublicKey}
	genesisBlock, err = BuildGenesisBlock(bookKeepers)
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
	_, pub := bls.GenerateRandomKey()
	block, err := BuildGenesisBlock([]bls.PublicKey{pub})
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.NotEqual(t, block.Header.TransactionsRoot, common.UINT256_EMPTY)
	assert.NotZero(t, len(block.Transactions))
	assert.NotNil(t, block.Transactions[0].Payload)
}

func TestSaveBridgeEventAsBlock(t *testing.T) {
	blockbefore := lg.GetCurrentBlockHash()
	var event = &payload.BridgeEvent{
		OriginData: wrappers.BridgeOracleRequest{
			RequestType: "setRequest",
			Bridge:      ethCommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
			Chainid:     big.NewInt(94),
		}}


	solEvent := payload.BridgeSolanaEvent{
		OriginData: wrappers.BridgeOracleRequestSolana{
			RequestType: "setRequest",
			Bridge:      [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 90, 1, 2, 3, 4, 5, 6, 7, 78, 9, 0, 1, 2, 2, 3, 43, 4, 4, 5, 5, 56, 23},
			Chainid:     big.NewInt(94),
		}}


	sol2EVMEvent := payload.SolanaToEVMEvent{
		OriginData: bridge.BridgeEvent{
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
		},
	}

	var be ledger.BlockEvents
	be.OracleRequests = append(be.OracleRequests, &event.OriginData)
	be.OracleSolanaRequests = append(be.OracleSolanaRequests, &solEvent.OriginData)
	be.SolanaBridgeEvents = append(be.SolanaBridgeEvents, &sol2EVMEvent.OriginData)
	_, err = lg.CreateBlockFromEvents(be)

	require.NoError(t, err)
	blockAfter := lg.GetCurrentBlockHash()
	require.Equal(t, blockbefore, blockAfter)

	t.Log("end")

}
