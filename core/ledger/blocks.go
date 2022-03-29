package ledger

import (
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/sirupsen/logrus"
)

func (l *Ledger) CreateBlockFromEvents(txs types.Transactions, sourceHeight uint64, epochBlockHash common.Uint256) (block *types.Block, err error) {
	prevHash := l.GetCurrentBlockHash()
	height := l.GetCurrentBlockHeight()
	block = types.NewBlock(l.GetChainId(), prevHash, epochBlockHash, sourceHeight, height+1, txs)
	return block, nil
}

func (l *Ledger) ExecAndSaveBlock(block *types.Block) error {
	result, err := l.ExecuteBlock(block)
	if err != nil {
		logrus.Error("ExecuteBlock")
		logrus.Error(err)
		return fmt.Errorf("execAndSaveBlock ExecuteBlock Height:%d error:%s", block.Header.Height, err)
	}
	err = l.SubmitBlock(block, result)
	if err != nil {
		logrus.Error("SubmitBlock")
		logrus.Error(err)
		return fmt.Errorf("execAndSaveBlock SubmitBlock Height:%d error:%s", block.Header.Height, err)
	}
	l.SetProcessedHeight(block.Header.SourceHeight)
	return nil
}
