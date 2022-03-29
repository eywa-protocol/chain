package ledgerstore

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"testing"

	"github.com/eywa-protocol/bls-crypto/bls"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/eywa-protocol/wrappers"
	"github.com/stretchr/testify/require"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/types"
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
	blockHeight := uint64(1)
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
	blockHeight := uint64(1)
	testBlockStore.NewBatch()
	testBlockStore.SaveBlockHash(blockHeight, blockHash)
	blockHash = sha256.Sum256([]byte("234567890"))
	blockHeight = uint64(2)
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
	invoke := &payload.InvokeCode{Code: []byte{1, 2, 3}}

	tx := types.ToTransaction(invoke)

	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	require.NoError(t, err)

	err = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	require.NoError(t, err)

	blockHeight := uint64(1)
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
	err = testBlockStore.SaveTransaction(tx.Payload, blockHeight)
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
	payload1, height, err := testBlockStore.GetTransaction(txHash)
	if err != nil {
		t.Errorf("GetTransaction error %s", err)
		return
	}
	if blockHeight != height {
		t.Errorf("TestSaveTransaction failed BlockHeight %d != %d", height, blockHeight)
		return
	}
	require.Equal(t, payload1.TxType(), tx.TxType(), "TestSaveTransaction failed TxType %d != %d", payload1.TxType, tx.TxType)

	tx1 := types.ToTransaction(payload1)
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

	tx := types.ToTransaction(event)

	sink := common.NewZeroCopySink(nil)

	err := tx.Serialization(sink)
	require.NoError(t, err)

	err = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	require.NoError(t, err)

	blockHeight := uint64(1)
	txHash := tx.Hash()
	exist, err := testBlockStore.ContainTransaction(txHash)
	require.NoError(t, err)
	require.False(t, exist)

	testBlockStore.NewBatch()
	err = testBlockStore.SaveTransaction(tx.Payload, blockHeight)
	require.NoError(t, err)

	err = testBlockStore.CommitTo()
	require.NoError(t, err)

	payload1, height, err := testBlockStore.GetTransaction(txHash)
	require.NoError(t, err)
	require.Equal(t, blockHeight, height)
	require.Equal(t, payload1.TxType(), tx.TxType())
	tx1 := types.ToTransaction(payload1)
	tx1Hash := tx1.Hash()
	require.Equal(t, txHash, tx1Hash)

	sink2 := common.NewZeroCopySink(nil)
	err = payload1.Serialization(sink2)
	require.NoError(t, err)
	var bridgeEvent2 payload.BridgeEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink2.Bytes()))
	require.NoError(t, err)
	require.Equal(t, tx.Payload, &bridgeEvent2)

	exist, err = testBlockStore.ContainTransaction(txHash)
	require.NoError(t, err)
	require.True(t, exist)
}

func TestSaveEpochTransaction(t *testing.T) {
	epoch := &payload.EpochEvent{
		Number:         0,
		EpochPublicKey: bls.PublicKey{},
		SourceTx:       common.Uint256{},
		PublicKeys:     nil,
	}
	tx := types.ToTransaction(epoch)

	sink := common.NewZeroCopySink(nil)

	err := tx.Serialization(sink)
	require.NoError(t, err)

	err = tx.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	require.NoError(t, err)

	blockHeight := uint64(1)
	txHash := tx.Hash()
	exist, err := testBlockStore.ContainTransaction(txHash)
	require.NoError(t, err)
	require.False(t, exist)

	testBlockStore.NewBatch()
	err = testBlockStore.SaveTransaction(tx.Payload, blockHeight)
	require.NoError(t, err)

	err = testBlockStore.CommitTo()
	require.NoError(t, err)

	payload1, height, err := testBlockStore.GetTransaction(txHash)
	require.NoError(t, err)
	require.Equal(t, blockHeight, height)
	require.Equal(t, payload1.TxType(), tx.TxType())
	tx1 := types.ToTransaction(payload1)
	tx1Hash := tx1.Hash()
	require.Equal(t, txHash, tx1Hash)

	sink2 := common.NewZeroCopySink(nil)
	err = payload1.Serialization(sink2)
	require.NoError(t, err)
	var ep2 payload.EpochEvent
	err = ep2.Deserialization(common.NewZeroCopySource(sink2.Bytes()))
	require.NoError(t, err)
	require.Equal(t, tx.Payload, &ep2)

	exist, err = testBlockStore.ContainTransaction(txHash)
	require.NoError(t, err)
	require.True(t, exist)

}

func TestHeaderIndexList(t *testing.T) {
	testBlockStore.NewBatch()
	startHeight := uint64(0)
	size := uint64(100)
	indexMap := make(map[uint64]common.Uint256, size)
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
	startHeight = uint64(100)
	size = uint64(100)
	indexMap = make(map[uint64]common.Uint256, size)
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
		h, ok := totalMap[uint64(height)]
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
	block := types.NewBlock(1111, common.Uint256{}, common.Uint256{}, 1, 1, types.Transactions{})
	blockHash := block.Hash()

	testBlockStore.NewBatch()

	err := testBlockStore.SaveHeader(block)
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

	if block.Header.Height != h.Height {
		t.Errorf("TestSaveHeader failed Height %d != %d", h.Height, block.Header.Height)
		return
	}
}

func TestBlock(t *testing.T) {
	pld := &payload.InvokeCode{}

	sink := common.NewZeroCopySink(nil)
	err := pld.Serialization(sink)
	if err != nil {
		t.Errorf("TestBlock SerializeUnsigned error:%s", err)
		return
	}
	_ = pld.Deserialization(common.NewZeroCopySource(sink.Bytes()))

	tx := types.ToTransaction(pld)
	block := types.NewBlock(1111, common.UINT256_EMPTY, common.UINT256_EMPTY, 2, 2, types.Transactions{tx})
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
	// t.Log(blockHash)
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
