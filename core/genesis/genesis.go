package genesis

import (
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/types"
)

// BuildGenesisBlock returns the genesis block with default consensus bookkeeper list
func BuildGenesisBlock(chainId uint64, genesisHeight uint64) (*types.Block, error) {
	return types.NewBlock(chainId, common.Uint256{}, common.Uint256{}, genesisHeight, 0, types.Transactions{}), nil
}
