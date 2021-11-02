package test


import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"testing"
)

func TestHash(t *testing.T) {

	var data []common.Uint256
	a1 := common.Uint256(sha256.Sum256([]byte("a")))
	a2 := common.Uint256(sha256.Sum256([]byte("b")))
	a3 := common.Uint256(sha256.Sum256([]byte("c")))
	a4 := common.Uint256(sha256.Sum256([]byte("d")))
	a5 := common.Uint256(sha256.Sum256([]byte("e")))
	data = append(data, a1)
	data = append(data, a2)
	data = append(data, a3)
	data = append(data, a4)
	data = append(data, a5)
	hash := common.ComputeMerkleRoot(data)
	assert.NotEqual(t, hash, common.UINT256_EMPTY)

}

const N = 120000

func BenchmarkComputeMerkleRoot(b *testing.B) {
	data := make([]common.Uint256, N)
	for i := range data {
		data[i] = common.Uint256(sha256.Sum256([]byte(fmt.Sprint(i))))
	}

	for i := 0; i < b.N; i++ {
		common.ComputeMerkleRoot(data)
	}
}

func BenchmarkComputeMerkleRootOld(b *testing.B) {
	data := make([]common.Uint256, N)
	for i := range data {
		data[i] = common.Uint256(sha256.Sum256([]byte(fmt.Sprint(i))))
	}

	for i := 0; i < b.N; i++ {
		computeMerkleRootOld(data)
	}
}

func TestComputeMerkleRoot(t *testing.T) {
	for n := 0; n < 100; n++ {
		data := make([]common.Uint256, n)
		for i := range data {
			data[i] = common.Uint256(sha256.Sum256([]byte(fmt.Sprint(i))))
		}

		h1 := computeMerkleRootOld(data)

		h2 := common.ComputeMerkleRoot(data)
		assert.Equal(t, h1, h2)
	}
}

func doubleSha256(s []common.Uint256) common.Uint256 {
	b := new(bytes.Buffer)
	for _, d := range s {
		d.Serialize(b)
	}
	temp := sha256.Sum256(b.Bytes())
	f := sha256.Sum256(temp[:])

	return common.Uint256(f)
}

type merkleTree struct {
	Depth uint
	Root  *merkleTreeNode
}

type merkleTreeNode struct {
	Hash  common.Uint256
	Left  *merkleTreeNode
	Right *merkleTreeNode
}

func (t *merkleTreeNode) IsLeaf() bool {
	return t.Left == nil && t.Right == nil
}

//use []Uint256 to create a new merkleTree
func newMerkleTree(hashes []common.Uint256) (*merkleTree, error) {
	if len(hashes) == 0 {
		return nil, errors.New("NewMerkleTree input no item error.")
	}
	var height uint

	height = 1
	nodes := generateLeaves(hashes)
	for len(nodes) > 1 {
		nodes = levelUp(nodes)
		height += 1
	}
	mt := &merkleTree{
		Root:  nodes[0],
		Depth: height,
	}
	return mt, nil

}

//Generate the leaves nodes
func generateLeaves(hashes []common.Uint256) []*merkleTreeNode {
	var leaves []*merkleTreeNode
	for _, d := range hashes {
		node := &merkleTreeNode{
			Hash: d,
		}
		leaves = append(leaves, node)
	}
	return leaves
}

//calc the next level's hash use double sha256
func levelUp(nodes []*merkleTreeNode) []*merkleTreeNode {
	var nextLevel []*merkleTreeNode
	for i := 0; i < len(nodes)/2; i++ {
		var data []common.Uint256
		data = append(data, nodes[i*2].Hash)
		data = append(data, nodes[i*2+1].Hash)
		hash := doubleSha256(data)
		node := &merkleTreeNode{
			Hash:  hash,
			Left:  nodes[i*2],
			Right: nodes[i*2+1],
		}
		nextLevel = append(nextLevel, node)
	}
	if len(nodes)%2 == 1 {
		var data []common.Uint256
		data = append(data, nodes[len(nodes)-1].Hash)
		data = append(data, nodes[len(nodes)-1].Hash)
		hash := doubleSha256(data)
		node := &merkleTreeNode{
			Hash:  hash,
			Left:  nodes[len(nodes)-1],
			Right: nodes[len(nodes)-1],
		}
		nextLevel = append(nextLevel, node)
	}
	return nextLevel
}

//input a []uint256, create a merkleTree & calc the root hash
func computeMerkleRootOld(hashes []common.Uint256) common.Uint256 {
	if len(hashes) == 0 {
		return common.Uint256{}
	}
	if len(hashes) == 1 {
		return hashes[0]
	}
	tree, _ := newMerkleTree(hashes)
	return tree.Root.Hash
}
