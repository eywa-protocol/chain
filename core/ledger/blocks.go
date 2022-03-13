package ledger

import (
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/common/log"
	"github.com/eywa-protocol/chain/core/types"
)

func (self *Ledger) CreateBlockFromEvents(txs types.Transactions, sourceHeight uint64, epochBlockHash common.Uint256) (block *types.Block, err error) {
	prevHash := self.GetCurrentBlockHash()
	height := self.GetCurrentBlockHeight()
	block = types.NewBlock(self.GetChainId(), prevHash, epochBlockHash, sourceHeight, height+1, txs)
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
