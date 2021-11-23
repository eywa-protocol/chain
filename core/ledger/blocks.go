package ledger

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/payload"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
	"gitlab.digiu.ai/blockchainlaboratory/wrappers"
	"time"
)

func (self *Ledger) CreateBlockFromEvent(evt wrappers.BridgeOracleRequest) error {
	event := &payload.BridgeEvent{OriginData: evt}
	txs := []*types.Transaction{}
	tx := &types.Transaction{
		TxType:  types.BridgeEvent,
		Payload: event,
	}
	txs = append(txs, tx)
	block, err := self.makeBlock(txs)
	if err != nil {
		return errors.New(fmt.Sprintf("makeBlock %v", err.Error()))
	}
	return self.execAndSaveBlock(block)
}

func (self *Ledger) makeBlock(transactions []*types.Transaction) (block *types.Block, err error) {
	prevHash := self.GetCurrentBlockHash()
	height := self.GetCurrentBlockHeight()
	txHash := []common.Uint256{}
	txRoot := common.ComputeMerkleRoot(txHash)
	logrus.Tracef("prevHash: %v prevHeigth: %d", prevHash, height)
	blockRoot := self.GetBlockRootWithPreBlockHashes(height+1, []common.Uint256{prevHash})
	mainEpochKey, err := self.GetEpochState()
	if err != nil {
		return &types.Block{}, err
	}

	header := &types.Header{
		PrevBlockHash:    prevHash,
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        uint32(time.Now().Unix()),
		Height:           height + 1,
		ConsensusData:    common.GetNonce(),
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

func (self *Ledger) execAndSaveBlock(block *types.Block) error {

	result, err := self.ExecuteBlock(block)
	if err != nil {
		return fmt.Errorf("execAndSaveBlock ExecuteBlock Height:%d error:%s", block.Header.Height, err)
	}
	err = self.SubmitBlock(block, result)
	if err != nil {
		return fmt.Errorf("execAndSaveBlock SubmitBlock Height:%d error:%s", block.Header.Height, err)
	}
	return nil
}
