package types

import (
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/merkle"
)

func CalculateMerkleRoot(hashes []common.Uint256) common.Uint256 {
	if len(hashes) == 0 {
		return common.Uint256{}
	}
	tree := merkle.NewTree(0, nil, merkle.NewMemHashStore())
	for _, hash := range hashes {
		tree.Append(hash.ToArray())
	}
	return tree.Root()
}
