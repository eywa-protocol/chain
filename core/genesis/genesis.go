package genesis

import (
	"time"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/types"
)

// BuildGenesisBlock returns the genesis block with default consensus bookkeeper list
func BuildGenesisBlock(chainId uint64, genesisHeight uint64) (*types.Block, error) {
	return types.NewBlock(
		chainId,
		common.Uint256{},
		common.Uint256{},
		genesisHeight,
		0,
		time.Date(2022, 7, 31, 0, 0, 0, 0, time.UTC),
		types.Transactions{},
	), nil
}
