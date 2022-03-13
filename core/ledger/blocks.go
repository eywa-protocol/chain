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

func newBridgeFromSolanaEventTransaction(evt bridge.BridgeEvent) (payload.Payload, error) {
	return &payload.SolanaToEVMEvent{OriginData: evt}, nil
}

func newBridgeEventTransaction(evt wrappers.BridgeOracleRequest) (payload.Payload, error) {
	return &payload.BridgeEvent{OriginData: evt}, nil
}

func newReceiveRequestTransaction(evt wrappers.BridgeReceiveRequest) (payload.Payload, error) {
	return &payload.ReceiveRequestEvent{OriginData: evt}, nil
}

func newBridgeSolanaEventTransaction(evt wrappers.BridgeOracleRequestSolana) (payload.Payload, error) {
	return &payload.BridgeSolanaEvent{OriginData: evt}, nil
}

func (self *Ledger) CreateBlockFromEvents(txs types.Transactions, sourceHeight uint64) (block *types.Block, err error) {
	block, err = self.makeBlock(txs, sourceHeight)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("CreateBlockFromEvents %v", err.Error()))
	}
	return block, nil
}

func (self *Ledger) makeBlock(transactions types.Transactions, sourceHeight uint64) (block *types.Block, err error) {
	prevHash := self.GetCurrentBlockHash()

	height := self.GetCurrentBlockHeight()
	hashes := make([]common.Uint256, 0, len(transactions))
	for _, tx := range transactions {
		hashes = append(hashes, tx.Hash())
	}
	txRoot := types.CalculateMerkleRoot(hashes)

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
