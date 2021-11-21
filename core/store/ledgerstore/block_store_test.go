package ledgerstore

import (
	"crypto/sha256"
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
	"math/big"
	"testing"
	"time"

	"github.com/eywa-protocol/bls-crypto/bls"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/account"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/payload"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
)

func TestVersion(t *testing.T) {
	testBlockStore.NewBatch()
	version := byte(1)
	err := testBlockStore.SaveVersion(version)
	if err != nil {
		t.Errorf("SaveVersion error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	v, err := testBlockStore.GetVersion()
	if err != nil {
		t.Errorf("GetVersion error %s", err)
		return
	}
	if version != v {
		t.Errorf("TestVersion failed version %d != %d", v, version)
		return
	}
}

func TestCurrentBlock(t *testing.T) {
	blockHash := common.Uint256(sha256.Sum256([]byte("123456789")))
	blockHeight := uint32(1)
	testBlockStore.NewBatch()
	err := testBlockStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		t.Errorf("SaveCurrentBlock error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	hash, height, err := testBlockStore.GetCurrentBlock()
	if hash != blockHash {
		t.Errorf("TestCurrentBlock BlockHash %x != %x", hash, blockHash)
		return
	}
	if height != blockHeight {
		t.Errorf("TestCurrentBlock BlockHeight %x != %x", height, blockHeight)
		return
	}
}

func TestBlockHash(t *testing.T) {
	blockHash := common.Uint256(sha256.Sum256([]byte("123456789")))
	blockHeight := uint32(1)
	testBlockStore.NewBatch()
	testBlockStore.SaveBlockHash(blockHeight, blockHash)
	blockHash = common.Uint256(sha256.Sum256([]byte("234567890")))
	blockHeight = uint32(2)
	testBlockStore.SaveBlockHash(blockHeight, blockHash)
	err := testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	hash, err := testBlockStore.GetBlockHash(blockHeight)
	if err != nil {
		t.Errorf("GetBlockHash error %s", err)
		return
	}
	if hash != blockHash {
		t.Errorf("TestBlockHash failed BlockHash %x != %x", hash, blockHash)
		return
	}
}

func TestSaveTransaction(t *testing.T) {
	invoke := &payload.InvokeCode{}

	tx := &types.Transaction{
		TxType:  types.Invoke,
		Payload: invoke,
	}
	sink := common.NewZeroCopySink(nil)
	err := tx.SerializeUnsigned(sink)
	if err != nil {
		t.Errorf("TestSaveTransaction SerializeUnsigned error:%s", err)
		return
	}
	_ = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))

	blockHeight := uint32(1)
	txHash := tx.Hash()
	t.Log(txHash)
	exist, err := testBlockStore.ContainTransaction(txHash)
	if err != nil {
		t.Errorf("ContainTransaction error %s", err)
		return
	}
	if exist {
		t.Errorf("TestSaveTransaction ContainTransaction should be false.")
		return
	}

	testBlockStore.NewBatch()
	t.Log(tx)
	t.Log(blockHeight)
	err = testBlockStore.SaveTransaction(tx, blockHeight)
	if err != nil {
		t.Errorf("SaveTransaction error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	t.Log("\n", txHash)
	tx1, height, err := testBlockStore.GetTransaction(txHash)
	if err != nil {
		t.Errorf("GetTransaction error %s", err)
		return
	}
	if blockHeight != height {
		t.Errorf("TestSaveTransaction failed BlockHeight %d != %d", height, blockHeight)
		return
	}
	if tx1.TxType != tx.TxType {
		t.Errorf("TestSaveTransaction failed TxType %d != %d", tx1.TxType, tx.TxType)
		return
	}
	tx1Hash := tx1.Hash()

	if txHash != tx1Hash {
		t.Errorf("TestSaveTransaction failed TxHash %x != %x", tx1Hash, txHash)
		return
	}

	exist, err = testBlockStore.ContainTransaction(txHash)
	if err != nil {
		t.Errorf("ContainTransaction error %s", err)
		return
	}
	if !exist {
		t.Errorf("TestSaveTransaction ContainTransaction should be true.")
		return
	}
}

func TestSaveBridgeEventTransaction(t *testing.T) {
	event := &payload.BridgeEvent{
		OriginData: wrappers.BridgeOracleRequest{
			RequestType: "setRequest",
			Bridge:      ethCommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
			Chainid:     big.NewInt(94),
		}}

	tx := &types.Transaction{
		TxType:  types.BridgeEvent,
		Payload: event,
	}
	sink := common.NewZeroCopySink(nil)

	err := tx.SerializeUnsigned(sink)
	require.NoError(t, err)

	err = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	require.Error(t, err)

	blockHeight := uint32(1)
	txHash := tx.Hash()
	exist, err := testBlockStore.ContainTransaction(txHash)
	require.NoError(t, err)
	require.False(t, exist)

	testBlockStore.NewBatch()
	err = testBlockStore.SaveTransaction(tx, blockHeight)
	require.NoError(t, err)

	err = testBlockStore.CommitTo()
	require.NoError(t, err)

	tx1, height, err := testBlockStore.GetTransaction(txHash)
	require.NoError(t, err)
	require.Equal(t, blockHeight, height)
	require.Equal(t, tx1.TxType, tx.TxType)
	tx1Hash := tx1.Hash()
	require.Equal(t, txHash, tx1Hash)

	sink2 := common.NewZeroCopySink(nil)
	tx1.Payload.Serialization(sink2)
	var bridgeEvent2 payload.BridgeEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink2.Bytes()))
	require.NoError(t, err)
	require.Equal(t, event, &bridgeEvent2)

	exist, err = testBlockStore.ContainTransaction(txHash)
	require.NoError(t, err)
	require.True(t, exist)

}

func TestSaveEpochTransaction(t *testing.T) {

	epoch := &payload.Epoch{Data: []byte("123456")}
	tx := &types.Transaction{
		TxType:  types.Epoch,
		Payload: epoch,
	}
	sink := common.NewZeroCopySink(nil)

	err := tx.SerializeUnsigned(sink)
	require.NoError(t, err)

	err = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	require.Error(t, err)

	blockHeight := uint32(1)
	txHash := tx.Hash()
	exist, err := testBlockStore.ContainTransaction(txHash)
	require.NoError(t, err)
	require.False(t, exist)

	testBlockStore.NewBatch()
	err = testBlockStore.SaveTransaction(tx, blockHeight)
	require.NoError(t, err)

	err = testBlockStore.CommitTo()
	require.NoError(t, err)

	tx1, height, err := testBlockStore.GetTransaction(txHash)
	require.NoError(t, err)
	require.Equal(t, blockHeight, height)
	require.Equal(t, tx1.TxType, tx.TxType)
	tx1Hash := tx1.Hash()
	require.Equal(t, txHash, tx1Hash)

	sink2 := common.NewZeroCopySink(nil)
	tx1.Payload.Serialization(sink2)
	var ep2 payload.Epoch
	err = ep2.Deserialization(common.NewZeroCopySource(sink2.Bytes()))
	require.NoError(t, err)
	require.Equal(t, epoch, &ep2)

	exist, err = testBlockStore.ContainTransaction(txHash)
	require.NoError(t, err)
	require.True(t, exist)

}

func TestHeaderIndexList(t *testing.T) {
	testBlockStore.NewBatch()
	startHeight := uint32(0)
	size := uint32(100)
	indexMap := make(map[uint32]common.Uint256, size)
	indexList := make([]common.Uint256, 0)
	for i := startHeight; i < size; i++ {
		hash := common.Uint256(sha256.Sum256([]byte(fmt.Sprintf("%v", i))))
		indexMap[i] = hash
		indexList = append(indexList, hash)
	}
	err := testBlockStore.SaveHeaderIndexList(startHeight, indexList)
	if err != nil {
		t.Errorf("SaveHeaderIndexList error %s", err)
		return
	}
	startHeight = uint32(100)
	size = uint32(100)
	indexMap = make(map[uint32]common.Uint256, size)
	for i := startHeight; i < size; i++ {
		hash := common.Uint256(sha256.Sum256([]byte(fmt.Sprintf("%v", i))))
		indexMap[i] = hash
		indexList = append(indexList, hash)
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	totalMap, err := testBlockStore.GetHeaderIndexList()
	if err != nil {
		t.Errorf("GetHeaderIndexList error %s", err)
		return
	}

	for height, hash := range indexList {
		h, ok := totalMap[uint32(height)]
		if !ok {
			t.Errorf("TestHeaderIndexList failed height:%d hash not exist", height)
			return
		}
		if hash != h {
			t.Errorf("TestHeaderIndexList failed height:%d hash %x != %x", height, hash, h)
			return
		}
	}
}

func TestSaveHeader(t *testing.T) {
	acc1 := account.NewAccount(0)
	acc2 := account.NewAccount(0)
	bookkeeper, err := types.AddressFromPubLeySlice([]bls.PublicKey{acc1.PublicKey, acc2.PublicKey})
	if err != nil {
		t.Errorf("AddressFromBookkeepers error %s", err)
		return
	}
	header := &types.Header{
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(1),
		ConsensusData:    123456789,
		NextBookkeeper:   bookkeeper,
	}
	block := &types.Block{
		Header:       header,
		Transactions: []*types.Transaction{},
	}
	blockHash := block.Hash()

	testBlockStore.NewBatch()

	err = testBlockStore.SaveHeader(block)
	if err != nil {
		t.Errorf("SaveHeader error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	h, err := testBlockStore.GetHeader(blockHash)
	if err != nil {
		t.Errorf("GetHeader error %s", err)
		return
	}

	headerHash := h.Hash()
	if blockHash != headerHash {
		t.Errorf("TestSaveHeader failed HeaderHash %x != %x", headerHash, blockHash)
		return
	}

	if header.Height != h.Height {
		t.Errorf("TestSaveHeader failed Height %d != %d", h.Height, header.Height)
		return
	}
}

func TestBlock(t *testing.T) {
	acc1 := account.NewAccount(0)
	acc2 := account.NewAccount(0)
	bookkeeper, err := types.AddressFromPubLeySlice([]bls.PublicKey{acc1.PublicKey, acc2.PublicKey})
	if err != nil {
		t.Errorf("AddressFromBookkeepers error %s", err)
		return
	}
	header := &types.Header{
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(2),
		ConsensusData:    1234567890,
		NextBookkeeper:   bookkeeper,
	}

	invoke := &payload.InvokeCode{}
	tx := &types.Transaction{
		TxType:  types.Invoke,
		Payload: invoke,
	}
	sink := common.NewZeroCopySink(nil)
	err = tx.SerializeUnsigned(sink)
	if err != nil {
		t.Errorf("TestBlock SerializeUnsigned error:%s", err)
		return
	}
	_ = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))

	t.Log(tx.Hash())

	if err != nil {
		t.Errorf("TestBlock transferTx error:%s", err)
		return
	}

	block := &types.Block{
		Header:       header,
		Transactions: []*types.Transaction{tx},
	}
	blockHash := block.Hash()
	tx1Hash := tx.Hash()

	testBlockStore.NewBatch()

	err = testBlockStore.SaveBlock(block)
	if err != nil {
		t.Errorf("SaveHeader error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	//t.Log(blockHash)
	b, err := testBlockStore.GetBlock(blockHash)
	if err != nil {
		t.Errorf("GetBlock error %s", err)
		return
	}

	hash := b.Hash()
	if hash != blockHash {
		t.Errorf("TestBlock failed BlockHash %x != %x ", hash, blockHash)
		return
	}
	exist, err := testBlockStore.ContainTransaction(tx1Hash)
	if err != nil {
		t.Errorf("ContainTransaction error %s", err)
		return
	}
	if !exist {
		t.Errorf("TestBlock failed transaction %x should exist", tx1Hash)
		return
	}

	if len(block.Transactions) != len(b.Transactions) {
		t.Errorf("TestBlock failed Transaction size %d != %d ", len(b.Transactions), len(block.Transactions))
		return
	}
	if b.Transactions[0].Hash() != tx1Hash {
		t.Errorf("\nTestBlock failed transaction1 hash %x != %x", b.Transactions[0].Hash(), tx1Hash)
		return
	}
}
