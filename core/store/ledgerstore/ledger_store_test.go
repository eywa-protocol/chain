package ledgerstore

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/account"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/common/log"
	"github.com/eywa-protocol/chain/core/genesis"
	"github.com/eywa-protocol/chain/core/types"
)

var testBlockStore *BlockStore
var testStateStore *StateStore
var testLedgerStore *LedgerStoreImp

func TestMain(m *testing.M) {
	log.InitLog(0)

	var err error
	testLedgerStore, err = NewLedgerStore("test/ledger")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewLedgerStore error %s\n", err)
		return
	}

	testBlockDir := "test/block"
	testBlockStore, err = NewBlockStore(testBlockDir, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewBlockStore error %s\n", err)
		return
	}
	testStateDir := "test/state"
	merklePath := "test/" + MerkleTreeStorePath
	testStateStore, err = NewStateStore(testStateDir, merklePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewStateStore error %s\n", err)
		return
	}
	m.Run()
	err = testLedgerStore.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testLedgerStore.Close error %s\n", err)
		return
	}
	err = testBlockStore.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testBlockStore.Close error %s\n", err)
		return
	}
	err = testStateStore.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testStateStore.Close error %s", err)
		return
	}
	err = os.RemoveAll("./test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.RemoveAll error %s\n", err)
		return
	}
	os.RemoveAll("ActorLog")
}

func TestInitLedgerStoreWithGenesisBlock(t *testing.T) {
	acc1 := account.NewAccount(1)
	acc2 := account.NewAccount(2)
	acc3 := account.NewAccount(3)
	acc4 := account.NewAccount(4)
	acc5 := account.NewAccount(5)
	acc6 := account.NewAccount(6)
	acc7 := account.NewAccount(7)

	bookkeepers := []bls.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey, acc4.PublicKey, acc5.PublicKey, acc6.PublicKey, acc7.PublicKey}
	bookkeeper, err := types.AddressFromPubLeySlice(bookkeepers)
	require.NoError(t, err)
	require.NotEqual(t, bookkeeper, common.Address{})

	block, err := genesis.BuildGenesisBlock(bookkeepers)
	require.NoError(t, err)
	//header := &types.Header{
	//	Version:          0,
	//	PrevBlockHash:    common.Uint256{},
	//	TransactionsRoot: common.Uint256{},
	//	Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
	//	Height:           uint32(0),
	//	ConsensusData:    1234567890,
	//	NextEpoch:   bookkeeper,
	//}
	//block.Header = header
	//block := &types.Block{
	//	Header:       header,
	//	Transactions: []*types.Transaction{},
	//}

	err = testLedgerStore.InitLedgerStoreWithGenesisBlock(block)
	if err != nil {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock error %s", err)
		return
	}

	curBlockHeight := testLedgerStore.GetCurrentBlockHeight()
	curBlockHash := testLedgerStore.GetCurrentBlockHash()
	if curBlockHeight != block.Header.Height {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed CurrentBlockHeight %d != %d", curBlockHeight, block.Header.Height)
		return
	}
	if curBlockHash != block.Hash() {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed CurrentBlockHash %x != %x", curBlockHash, block.Hash())
		return
	}
	block1, err := testLedgerStore.GetBlockByHeight(curBlockHeight)
	if err != nil {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed GetBlockByHeight error %s", err)
		return
	}

	if block1.Hash() != block.Hash() {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed blockhash %x != %x", block1.Hash(), block.Hash())
		return
	}

	blockByHash, err := testLedgerStore.GetBlockByHash(curBlockHash)
	if err != nil {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed GetBlockByHash error %s", err)
		return
	}

	if blockByHash.Hash() != block.Hash() {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed blockhash %x != %x", blockByHash.Hash(), block.Hash())
		return
	}
}
