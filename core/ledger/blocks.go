package ledger

import (
	"errors"
	"fmt"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/common/log"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/eywa-protocol/wrappers"
)

func newBridgeEventTransaction(evt wrappers.BridgeOracleRequest) (*types.Transaction, error) {
	event := &payload.BridgeEvent{OriginData: evt}
	tx := &types.Transaction{
		TxType:  types.BridgeEvent,
		Payload: event,
		ChainID: 0,
	}
	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	if err != nil {
		return &types.Transaction{}, err
	}
	tx, err = types.TransactionFromRawBytes(sink.Bytes())
	if err != nil {
		return &types.Transaction{}, err
	}
	return tx, nil
}

func newBridgeSolanaEventTransaction(evt wrappers.BridgeOracleRequestSolana) (*types.Transaction, error) {
	event := &payload.BridgeSolanaEvent{OriginData: evt}
	tx := &types.Transaction{
		TxType:  types.BridgeEventSolana,
		Payload: event,
		ChainID: 0,
	}
	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	if err != nil {
		return &types.Transaction{}, err
	}
	tx, err = types.TransactionFromRawBytes(sink.Bytes())
	if err != nil {
		return &types.Transaction{}, err
	}
	return tx, nil
}

func (self *Ledger) CreateBlockFromEvent(evt wrappers.BridgeOracleRequest) (block *types.Block, err error) {
	txs := []*types.Transaction{}
	tx, err := newBridgeEventTransaction(evt)
	if err != nil {
		return nil, err
	}
	txs = append(txs, tx)
	block, err = self.makeBlock(txs)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("makeBlock %v", err.Error()))
	}
	return block, self.ExecAndSaveBlock(block)
}

func (self *Ledger) CreateBlockFromSolanaEvent(evt wrappers.BridgeOracleRequestSolana) (block *types.Block, err error) {
	txs := []*types.Transaction{}
	tx, err := newBridgeSolanaEventTransaction(evt)
	if err != nil {
		return nil, err
	}
	txs = append(txs, tx)
	block, err = self.makeBlock(txs)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("makeBlock %v", err.Error()))
	}
	return block, self.ExecAndSaveBlock(block)
}

func (self *Ledger) makeBlock(transactions []*types.Transaction) (block *types.Block, err error) {
	prevHash := self.GetCurrentBlockHash()

	height := self.GetCurrentBlockHeight()
	hashes := make([]common.Uint256, 0, len(transactions))
	for _, tx := range transactions {
		hashes = append(hashes, tx.Hash())
	}
	txRoot := common.ComputeMerkleRoot(hashes)
	blockRoot := self.GetBlockRootWithPreBlockHashes(height+1, []common.Uint256{prevHash})

	mainEpochKey, err := self.GetEpochState()
	if err != nil {
		return &types.Block{}, err
	}
	log.Infof(" - > prev hash %v", prevHash.ToHexString())
	log.Infof(" - > blockRoot %v", blockRoot.ToHexString())
	log.Infof(" - > height %v", height)
	log.Infof(" - > txRoot %v", txRoot.ToHexString())

	header := &types.Header{
		PrevBlockHash:    prevHash,
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        transactions[0].Nonce,
		Height:           height + 1,
		ConsensusData:    uint64(transactions[0].Nonce),
		EpochKey:         mainEpochKey.CurrEpoch[0],
	}
	block = &types.Block{
		Header:       header,
		Transactions: transactions,
	}

	//blockHash := block.Hash()

	//sig := self.PrivKey.Sign(blockHash[:])

	//block.Header.SigData = []bls.Signature{sig}
	return block, nil
}

func (self *Ledger) ExecAndSaveBlock(block *types.Block) error {
	result, err := self.ExecuteBlock(block)
	if err != nil {
		log.Error("ExecuteBlock")
		log.Error(err)
		return fmt.Errorf("execAndSaveBlock ExecuteBlock Height:%d error:%s", block.Header.Height, err)
	}
	err = self.SubmitBlock(block, result)
	if err != nil {
		log.Error("SubmitBlock")
		log.Error(err)
		return fmt.Errorf("execAndSaveBlock SubmitBlock Height:%d error:%s", block.Header.Height, err)
	}
	return nil
}
