package ledgerstore

import (
	"math/rand"
	"testing"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/merkle"
	"github.com/stretchr/testify/assert"
)

func TestStateMerkleRoot(t *testing.T) {
	teststatemerkleroot := func(H, effectiveStateHashHeight uint64) {
		diffHashes := make([]common.Uint256, 0, H)
		for i := uint64(0); i < H; i++ {
			var hash common.Uint256
			rand.Read(hash[:])
			diffHashes = append(diffHashes, hash)
		}
		db := NewMemStateStore(effectiveStateHashHeight)
		for h, hash := range diffHashes[:effectiveStateHashHeight] {
			height := uint64(h)
			db.NewBatch()
			err := db.AddStateMerkleTreeRoot(height, hash)
			assert.Nil(t, err)
			db.CommitTo()
			root, _ := db.GetStateMerkleRoot(height)
			assert.Equal(t, root, common.UINT256_EMPTY)
		}

		merkleTree := merkle.NewTree(0, nil, nil)
		for h, hash := range diffHashes[effectiveStateHashHeight:] {
			height := uint64(h) + effectiveStateHashHeight
			merkleTree.Append(hash.ToArray())
			root1 := db.GetStateMerkleRootWithNewHash(hash)
			db.NewBatch()
			err := db.AddStateMerkleTreeRoot(height, hash)
			assert.Nil(t, err)
			db.CommitTo()
			root2, _ := db.GetStateMerkleRoot(height)
			root3 := merkleTree.Root()

			assert.Equal(t, root1, root2)
			assert.Equal(t, root1, root3)
		}
	}

	for i := 0; i < 200; i++ {
		teststatemerkleroot(1024, uint64(i))
		h := rand.Uint64()%1000 + 1
		eff := rand.Uint64() % h
		teststatemerkleroot(h, eff)
	}

}
