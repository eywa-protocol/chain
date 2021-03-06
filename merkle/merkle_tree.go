package merkle

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/eywa-protocol/chain/common"
	"github.com/sirupsen/logrus"
)

// TODO: fix unhandled errors

// const UINT256_SIZE int = 64

// type common.Uint256 [UINT256_SIZE]byte

var EMPTY_HASH = common.Uint256{}

// CompactMerkleTree calculate merkle tree with compact hash store in HashStore
type CompactMerkleTree struct {
	mintree_h uint
	hashes    []common.Uint256
	hasher    TreeHasher
	hashStore HashStore
	rootHash  common.Uint256
	treeSize  uint64
}

// NewTree returns a CompactMerkleTree instance
func NewTree(tree_size uint64, hashes []common.Uint256, store HashStore) *CompactMerkleTree {

	tree := &CompactMerkleTree{
		mintree_h: 0,
		hashes:    nil,
		hasher:    TreeHasher{},
		hashStore: store,
		rootHash:  EMPTY_HASH,
	}

	tree._update(tree_size, hashes)
	return tree
}

func (t *CompactMerkleTree) Hashes() []common.Uint256 {
	return t.hashes
}

func (t *CompactMerkleTree) TreeSize() uint64 {
	return t.treeSize
}

func (t *CompactMerkleTree) Marshal() ([]byte, error) {
	length := 8 + len(t.hashes)*common.UINT256_SIZE
	buf := make([]byte, 8, length)
	binary.BigEndian.PutUint64(buf[0:], t.treeSize)
	for _, h := range t.hashes {
		buf = append(buf, h[:]...)
	}

	return buf, nil
}

func (t *CompactMerkleTree) UnMarshal(buf []byte) error {
	tree_size := binary.BigEndian.Uint64(buf[0:8])
	nhashes := countBit(tree_size)
	if len(buf) < 8+int(nhashes)*common.UINT256_SIZE {
		return errors.New("Too short input buf length")
	}
	hashes := make([]common.Uint256, nhashes, nhashes)
	for i := 0; i < int(nhashes); i++ {
		copy(hashes[i][:], buf[8+i*common.UINT256_SIZE:])
	}

	t._update(tree_size, hashes)

	return nil
}

func (t *CompactMerkleTree) _update(tree_size uint64, hashes []common.Uint256) {
	numBit := countBit(tree_size)
	if len(hashes) != int(numBit) {
		panic("number of hashes != num bit in tree_size")
	}
	t.treeSize = tree_size
	t.hashes = hashes
	t.mintree_h = lowBit(tree_size)
	t.rootHash = EMPTY_HASH

}

// Root returns root hash of merkle tree
func (t *CompactMerkleTree) Root() common.Uint256 {
	if t.rootHash == EMPTY_HASH {
		if len(t.hashes) != 0 {
			t.rootHash = t.hasher._hash_fold(t.hashes)
		} else {
			t.rootHash = t.hasher.hash_empty()
		}
	}
	return t.rootHash
}

// GetRootWithNewLeaf returns the new root hash if newLeaf is appended to the merkle tree
func (t *CompactMerkleTree) GetRootWithNewLeaf(newLeaf common.Uint256) common.Uint256 {
	hashes := append(t.hashes, t.hasher.hash_leaf(newLeaf.ToArray()))
	root := t.hasher._hash_fold(hashes)

	return root
}

// clone except internal hash storage
func (t *CompactMerkleTree) cloneMem() CompactMerkleTree {
	temp := CompactMerkleTree{mintree_h: t.mintree_h, hasher: t.hasher, hashStore: nil,
		rootHash: t.rootHash, treeSize: t.treeSize,
	}
	temp.hashes = make([]common.Uint256, len(t.hashes))
	for i, h := range t.hashes {
		temp.hashes[i] = h
	}

	return temp
}

func (t *CompactMerkleTree) GetRootWithNewLeaves(newLeaf []common.Uint256) common.Uint256 {
	tree := t.cloneMem()
	for _, h := range newLeaf {
		tree.Append(h.ToArray())
	}

	return tree.Root()
}

// Append appends a leaf to the merkle tree and returns the audit path
func (t *CompactMerkleTree) Append(leafv []byte) []common.Uint256 {
	leaf := t.hasher.hash_leaf(leafv)

	return t.appendHash(leaf)
}

// AppendHash appends a leaf hash to the merkle tree and returns the audit path
func (t *CompactMerkleTree) appendHash(leaf common.Uint256) []common.Uint256 {
	size := len(t.hashes)
	auditPath := make([]common.Uint256, size, size)
	storehashes := make([]common.Uint256, 0)
	// reverse
	for i, v := range t.hashes {
		auditPath[size-i-1] = v
	}

	storehashes = append(storehashes, leaf)
	t.mintree_h = 1
	for s := t.treeSize; s%2 == 1; s = s >> 1 {
		t.mintree_h += 1
		leaf = t.hasher.hash_children(t.hashes[size-1], leaf)
		storehashes = append(storehashes, leaf)
		size -= 1
	}
	if t.hashStore != nil {
		t.hashStore.Append(storehashes)
		t.hashStore.Flush()
	}
	t.treeSize += 1
	t.hashes = t.hashes[0:size]
	t.hashes = append(t.hashes, leaf)
	t.rootHash = EMPTY_HASH

	return auditPath
}

func (t *CompactMerkleTree) DumpStatus() {
	logrus.Errorf("tree root: %x \n", t.rootHash)
	logrus.Errorf("tree size: %d \n", t.treeSize)
	logrus.Errorf("hashes size: %d \n", len(t.hashes))
	logrus.Errorf("hashes  %x \n", t.hashes)
	logrus.Errorf("mintree_h  %x \n", t.mintree_h)
}

// 1 based n
func getSubTreeSize(n uint64) []uint64 {
	nhashes := countBit(n)
	subtreesize := make([]uint64, nhashes, nhashes)
	for i, id := nhashes-1, uint64(1); n != 0; n = n >> 1 {
		id = id * 2
		if n%2 == 1 {
			subtreesize[i] = id - 1
			i -= 1
		}
	}

	return subtreesize
}

// 1-based n and return value
func getSubTreePos(n uint64) []uint64 {
	nhashes := countBit(n)
	hashespos := make([]uint64, nhashes, nhashes)
	for i, id := nhashes-1, uint64(1); n != 0; n = n >> 1 {
		id = id * 2
		if n%2 == 1 {
			hashespos[i] = id - 1
			i -= 1
		}
	}

	for i := uint(1); i < nhashes; i++ {
		hashespos[i] += hashespos[i-1]
	}

	return hashespos
}

// return merkle root of D[0:n] not include n
func (t *CompactMerkleTree) merkleRoot(n uint64) common.Uint256 {
	hashespos := getSubTreePos(n)
	nhashes := uint(len(hashespos))

	hashes := make([]common.Uint256, nhashes, nhashes)
	for i := uint(0); i < nhashes; i++ {
		hashes[i], _ = t.hashStore.GetHash(hashespos[i] - 1)
	}
	return t.hasher._hash_fold(hashes)
}

// ConsistencyProof returns consistency proof
func (t *CompactMerkleTree) ConsistencyProof(m, n uint64) []common.Uint256 {
	if m > n || t.treeSize < n || t.hashStore == nil {
		return nil
	}

	return t.subproof(m, n, true)
}

// m, n 1-based
func (t *CompactMerkleTree) subproof(m, n uint64, b bool) []common.Uint256 {
	offset := uint64(0)
	var hashes []common.Uint256
	for m < n {
		k := uint64(1 << (highBit(n-1) - 1))
		if m <= k {
			pos := getSubTreePos(n - k)
			subhashes := make([]common.Uint256, len(pos), len(pos))
			for p := range pos {
				pos[p] += offset + k*2 - 1
				subhashes[p], _ = t.hashStore.GetHash(pos[p] - 1)
			}
			rootk2n := t.hasher._hash_fold(subhashes)
			hashes = append(hashes, rootk2n)
			n = k
		} else {
			offset += k*2 - 1
			root02k, _ := t.hashStore.GetHash(offset - 1)
			hashes = append(hashes, root02k)
			m -= k
			n -= k
			b = false
		}
	}

	// assert m == n
	if b == false {
		pos := getSubTreePos(n)
		// assert len(pos) == 1
		if len(pos) != 1 {
			panic("assert error")
		}
		root02n, _ := t.hashStore.GetHash(pos[0] + offset - 1)
		hashes = append(hashes, root02n)
	}

	length := len(hashes)
	reverse := make([]common.Uint256, length, length)
	for k := range reverse {
		reverse[k] = hashes[length-k-1]
	}

	return reverse
}

// InclusionProof returns the proof d[m] in D[0:n]
// m zero based index, n size 1-based
// return sink.Bytes() of WriteVarBytes(hash_index_by_m) + loop of { WriteByte(PosInfo) + WriteByte(ProofPathNodeHash) }
func (t *CompactMerkleTree) MerkleInclusionLeafPath(data []byte, m, n uint64) ([]byte, error) {
	if m >= n {
		return nil, errors.New("wrong parameters")
	} else if t.treeSize < n {
		return nil, errors.New("not available yet")
	} else if t.hashStore == nil {
		return nil, errors.New("hash store not available")
	}

	offset := uint64(0)
	depth := int(math.Ceil(math.Log2(float64(n))))
	hashes := make([]common.Uint256, 0, depth)
	poses := make([]byte, 0, depth)
	for n != 1 {
		k := uint64(1 << (highBit(n-1) - 1))
		if m < k {
			pos := getSubTreePos(n - k)
			subhashes := make([]common.Uint256, len(pos), len(pos))
			for p := range pos {
				pos[p] += offset + k*2 - 1
				subhashes[p], _ = t.hashStore.GetHash(pos[p] - 1)
			}
			rootk2n := t.hasher._hash_fold(subhashes)
			hashes = append(hashes, rootk2n)
			poses = append(poses, byte(1))
			n = k
		} else {
			offset += k*2 - 1
			root02k, _ := t.hashStore.GetHash(offset - 1)
			hashes = append(hashes, root02k)
			poses = append(poses, byte(0))
			m -= k
			n -= k
		}
	}
	length := len(hashes)
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes(data)
	for k := range hashes {
		index := length - k - 1
		sink.WriteByte(poses[index])
		sink.WriteHash(hashes[index])
	}
	return sink.Bytes(), nil
}

// InclusionProof returns the proof d[m] in D[0:n]
// m zero based index, n size 1-based
func (t *CompactMerkleTree) InclusionProof(m, n uint64) ([]common.Uint256, error) {
	if m >= n {
		return nil, errors.New("wrong parameters")
	} else if t.treeSize < n {
		return nil, errors.New("not available yet")
	} else if t.hashStore == nil {
		return nil, errors.New("hash store not available")
	}

	offset := uint64(0)
	var hashes []common.Uint256
	for n != 1 {
		k := uint64(1 << (highBit(n-1) - 1))
		if m < k {
			pos := getSubTreePos(n - k)
			subhashes := make([]common.Uint256, len(pos), len(pos))
			for p := range pos {
				pos[p] += offset + k*2 - 1
				subhashes[p], _ = t.hashStore.GetHash(pos[p] - 1)
			}
			rootk2n := t.hasher._hash_fold(subhashes)
			hashes = append(hashes, rootk2n)
			n = k
		} else {
			offset += k*2 - 1
			root02k, _ := t.hashStore.GetHash(offset - 1)
			hashes = append(hashes, root02k)
			m -= k
			n -= k
		}
	}

	length := len(hashes)
	reverse := make([]common.Uint256, length, length)
	for k := range reverse {
		reverse[k] = hashes[length-k-1]
	}

	return reverse, nil
}

// MerkleVerifier verify inclusion and consist proof
type MerkleVerifier struct {
	hasher TreeHasher
}

func NewMerkleVerifier() *MerkleVerifier {
	return &MerkleVerifier{
		hasher: TreeHasher{},
	}
}

/*
   Verify a Merkle Audit PATH.

   leaf_hash: The hash of the leaf for which the proof was provided.
   leaf_index: Index of the leaf in the tree.
   proof: A list of SHA-256 hashes representing the  Merkle audit path.
   tree_size: The size of the tree
   root_hash: The root hash of the tree

   Returns:
       nil when the proof is valid
*/
func (self *MerkleVerifier) VerifyLeafHashInclusion(leaf_hash common.Uint256,
	leaf_index uint64, proof []common.Uint256, root_hash common.Uint256, tree_size uint64) error {

	if tree_size <= leaf_index {
		return errors.New("Wrong params: the tree size is smaller than the leaf index")
	}

	calculated_root_hash, err := self.calculate_root_hash_from_audit_path(leaf_hash,
		leaf_index, proof, tree_size)
	if err != nil {
		return err
	}
	if calculated_root_hash != root_hash {
		return errors.New(fmt.Sprintf("Constructed root hash differs from provided root hash. Constructed: %x, Expected: %x",
			calculated_root_hash, root_hash))
	}

	return nil
}

/*
   Verify a Merkle Audit PATH.

   leaf: The leaf for which the proof is provided
   leaf_index: Index of the leaf in the tree.
   proof: A list of SHA-256 hashes representing the  Merkle audit path.
   tree_size: The size of the tree
   root_hash: The root hash of the tree

   Returns:
       nil when the proof is valid
*/
func (self *MerkleVerifier) VerifyLeafInclusion(leaf []byte,
	leaf_index uint64, proof []common.Uint256, root_hash common.Uint256, tree_size uint64) error {
	leaf_hash := self.hasher.hash_leaf(leaf)
	return self.VerifyLeafHashInclusion(leaf_hash, leaf_index, proof, root_hash, tree_size)
}

func (self *MerkleVerifier) calculate_root_hash_from_audit_path(leaf_hash common.Uint256,
	node_index uint64, audit_path []common.Uint256, tree_size uint64) (common.Uint256, error) {
	calculated_hash := leaf_hash
	last_node := tree_size - 1
	pos := 0
	path_len := len(audit_path)
	for last_node > 0 {
		if pos >= path_len {
			return EMPTY_HASH, errors.New(fmt.Sprintf("Proof too short. expected %d, got %d",
				audit_path_length(node_index, tree_size), path_len))
		}

		if node_index%2 == 1 {
			calculated_hash = self.hasher.hash_children(audit_path[pos], calculated_hash)
			pos += 1
		} else if node_index < last_node {
			calculated_hash = self.hasher.hash_children(calculated_hash, audit_path[pos])
			pos += 1
		}
		node_index /= 2
		last_node /= 2
	}

	if pos < path_len {
		return EMPTY_HASH, errors.New("Proof too long")
	}
	return calculated_hash, nil
}

func audit_path_length(index, tree_size uint64) int {
	length := 0
	last_node := tree_size - 1
	for last_node > 0 {
		if index%2 == 1 || index < last_node {
			length += 1
		}
		index /= 2
		last_node /= 2
	}
	return length
}

/*
Verify the consistency between two root hashes.

    old_tree_size must be <= new_tree_size.

    Args:
        old_tree_size: size of the older tree.
        new_tree_size: size of the newer_tree.
        old_root: the root hash of the older tree.
        new_root: the root hash of the newer tree.
        proof: the consistency proof.

    Returns:
        True. The return value is enforced by a decorator and need not be
            checked by the caller.

    Raises:
        ConsistencyError: the proof indicates an inconsistency
            (this is usually really serious!).
        ProofError: the proof is invalid.
        ValueError: supplied tree sizes are invalid.
*/
func (self *MerkleVerifier) VerifyConsistency(old_tree_size,
	new_tree_size uint64, old_root, new_root common.Uint256, proof []common.Uint256) error {
	old_size := old_tree_size
	new_size := new_tree_size

	if old_size > new_size {
		return errors.New(fmt.Sprintf("Older tree has bigger size %d vs %d", old_size, new_size))
	}
	if old_root == new_root {
		return nil
	}
	if old_size == 0 {
		return nil
	}
	// assert o < old_size < new_size
	/*
		A consistency proof is essentially an audit proof for the node with
		index old_size - 1 in the newer tree. The sole difference is that
		the path is already hashed together into a single hash up until the
		first audit node that occurs in the newer tree only.
	*/
	node := old_size - 1
	last_node := new_size - 1

	// while we are the right child, everything is in both trees, so move one level up
	for node%2 == 1 {
		node /= 2
		last_node /= 2
	}

	lenp := len(proof)
	pos := 0
	var new_hash, old_hash common.Uint256

	if pos >= lenp {
		return errors.New("Wrong proof length")
	}
	if node != 0 {
		// compute the two root hashes in parallel.
		new_hash = proof[pos]
		old_hash = proof[pos]
		pos += 1
	} else {
		// The old tree was balanced (2^k nodes), so we already have the first root hash
		new_hash = old_root
		old_hash = old_root
	}

	for node != 0 {
		if node%2 == 1 {
			if pos >= lenp {
				return errors.New("Wrong proof length")
			}
			// node is a right child: left sibling exists in both trees
			next_node := proof[pos]
			pos += 1
			old_hash = self.hasher.hash_children(next_node, old_hash)
			new_hash = self.hasher.hash_children(next_node, new_hash)
		} else if node < last_node {
			if pos >= lenp {
				return errors.New("Wrong proof length")
			}
			// node is a left child: right sibling only exists inthe newer tree
			next_node := proof[pos]
			pos += 1
			new_hash = self.hasher.hash_children(new_hash, next_node)
		}
		// else node == last_node: node is a left child with no sibling in either tree

		node /= 2
		last_node /= 2
	}

	// Now old_hash is the hash of the first subtree. If the two trees have different
	// height, continue the path until the new root.
	for last_node != 0 {
		if pos >= lenp {
			return errors.New("Wrong proof length")
		}
		next_node := proof[pos]
		pos += 1
		new_hash = self.hasher.hash_children(new_hash, next_node)
		last_node /= 2
	}

	/* If the second hash does not match, the proof is invalid for the given pair
	If, on the other hand, the newer hash matches but the older one does not, then
	the proof (together with the signatures on the hashes) is proof of inconsistency.
	*/
	if new_hash != new_root {
		return errors.New(fmt.Sprintf(`Bad Merkle proof: second root hash does not match. 
			Expected hash:%x, computed hash: %x`, new_root, new_hash))
	} else if old_hash != old_root {
		return errors.New(fmt.Sprintf(`Inconsistency: first root hash does not match."
			"Expected hash: %x, computed hash:%x`, old_root, old_hash))
	}

	if pos != lenp {
		return errors.New("Proof too long")
	}

	return nil
}
