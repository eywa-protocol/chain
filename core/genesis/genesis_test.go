package genesis

import (
	"fmt"
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/account"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/cmd/utils"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/ledger"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
	"io/ioutil"
	"os"
	"testing"
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
	saveBlockToFile(genesisBlock, genesisFile)
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

func saveBlockToFile(block *types.Block, file string) {

	f, err := os.Create(file)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	genesisBlockBytes := block.ToArray()
	err = ioutil.WriteFile(file, genesisBlockBytes, os.ModePerm)
	defer f.Close()
	if err != nil {
		panic(err)
	}
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

	t.Log("bookkeepers", genesisBlockFromBytes.Header.EpochValidators[0].Marshal())

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
