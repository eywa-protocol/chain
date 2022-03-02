package ledger

import (
	"errors"
	"fmt"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/common/log"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/eywa-protocol/wrappers"
)

type BlockEvents struct {
	OracleRequests       []*wrappers.BridgeOracleRequest
	ReceiveRequests      []*wrappers.BridgeReceiveRequest
	OracleSolanaRequests []*wrappers.BridgeOracleRequestSolana
	SolanaBridgeEvents   []*bridge.BridgeEvent
}

func newBridgeFromSolanaEventTransaction(evt bridge.BridgeEvent) (*types.Transaction, error) {
	event := &payload.SolanaToEVMEvent{OriginData: evt}
	tx := &types.Transaction{
		TxType:  types.SolanaToEVMEvent,
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

func newReceiveRequestTransaction(evt wrappers.BridgeReceiveRequest) (*types.Transaction, error) {
	event := &payload.ReceiveRequestEvent{OriginData: evt}
	tx := &types.Transaction{
		TxType:  types.ReceiveRequestEvent,
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

func (self *Ledger) CreateBlockFromEvents(blockEvents BlockEvents, sourceHeight uint64) (block *types.Block, err error) {

	txs := []*types.Transaction{}
	for _, tx1 := range blockEvents.OracleRequests {
		tx, err := newBridgeEventTransaction(*tx1)
		if err != nil {
			log.Errorf("newBridgeEventTransaction: %v", err)
			continue
		}
		txs = append(txs, tx)
	}

	for _, tx1 := range blockEvents.OracleSolanaRequests {
		tx, err := newBridgeSolanaEventTransaction(*tx1)
		if err != nil {
			log.Errorf("newBridgeSolanaEventTransaction: %v", err)
			continue
		}
		txs = append(txs, tx)
	}

	for _, tx1 := range blockEvents.ReceiveRequests {
		tx, err := newReceiveRequestTransaction(*tx1)
		if err != nil {
			log.Errorf("newReceiveRequestTransaction: %v", err)
			continue
		}
		txs = append(txs, tx)
	}

	for _, tx1 := range blockEvents.SolanaBridgeEvents {
		tx, err := newBridgeFromSolanaEventTransaction(*tx1)
		if err != nil {
			log.Errorf("newReceiveRequestTransaction: %v", err)
			continue
		}
		txs = append(txs, tx)
	}

	block, err = self.makeBlock(txs, sourceHeight)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("CreateBlockFromEvents %v", err.Error()))
	}
	return block, nil
}

func (self *Ledger) makeBlock(transactions []*types.Transaction, sourceHeight uint64) (block *types.Block, err error) {
	prevHash := self.GetCurrentBlockHash()

	height := self.GetCurrentBlockHeight()
	hashes := make([]common.Uint256, 0, len(transactions))
	for _, tx := range transactions {
		hashes = append(hashes, tx.Hash())
	}
	txRoot := common.ComputeMerkleRoot(hashes)
	blockRoot := self.GetBlockRootWithPreBlockHashes(height+1, []common.Uint256{prevHash})

	log.Infof(" - > prev hash %v", prevHash.ToHexString())
	log.Infof(" - > blockRoot %v", blockRoot.ToHexString())
	log.Infof(" - > height %v", height)
	log.Infof(" - > txRoot %v", txRoot.ToHexString())

	header := &types.Header{
		PrevBlockHash:    prevHash,
		TransactionsRoot: txRoot,
		Height:           height + 1,
		SourceHeight:     sourceHeight,
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
