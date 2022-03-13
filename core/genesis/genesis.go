package genesis

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/types"
)

const (
	INIT_CONFIG = "initConfig"
)

// BuildGenesisBlock returns the genesis block with default consensus bookkeeper list
func BuildGenesisBlock(chainId uint64) (*types.Block, error) {
	genesisHeader := &types.Header{
		ChainID:          chainId,
		PrevBlockHash:    common.Uint256{},
		EpochBlockHash:   common.Uint256{},
		TransactionsRoot: common.Uint256{},
		SourceHeight:     0,
		Height:           0,
		Signature:        bls.NewZeroMultisig(),
	}

	genesisBlock := &types.Block{
		Header:       genesisHeader,
		Transactions: types.Transactions{},
	}
	genesisBlock.RebuildMerkleRoot()
	return genesisBlock, nil
}
