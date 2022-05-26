package types

import (
	"errors"
	"fmt"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/merkle"
)

type Block struct {
	Header       *Header
	Transactions Transactions

	merkleTree *merkle.CompactMerkleTree
}

func (b *Block) Serialization(sink *common.ZeroCopySink) error {
	err := b.Header.Serialization(sink)
	if err != nil {
		return err
	}

	return b.Transactions.Serialization(sink)
}

func NewBlock(chainId uint64, prevHash common.Uint256, epochBlockHash common.Uint256, sourceHeight uint64, height uint64, transactions Transactions) *Block {
	header := &Header{
		ChainID:        chainId,
		PrevBlockHash:  prevHash,
		EpochBlockHash: epochBlockHash,
		SourceHeight:   sourceHeight,
		Height:         height,
		Signature:      bls.NewZeroMultisig(),
	}
	return NewBlockFromComponents(header, transactions)
}

func NewBlockFromComponents(header *Header, transactions Transactions) *Block {
	block := &Block{
		Header:       header,
		Transactions: transactions,
	}
	block.rebuildMerkleRoot()
	block.Header.сalculateHash()

	return block
}

func BlockFromRawBytes(raw []byte) (*Block, error) {
	source := common.NewZeroCopySource(raw)
	block := &Block{}
	err := block.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (b *Block) Deserialization(source *common.ZeroCopySource) error {
	if b.Header == nil {
		b.Header = new(Header)
	}
	if err := b.Header.Deserialization(source); err != nil {
		return err
	}

	if err := b.Transactions.Deserialization(source); err != nil {
		return err
	}
	b.rebuildMerkleRoot()
	b.Header.сalculateHash()
	return nil
}

func (b *Block) VerifyIntegrity() error {
	mask := make(map[common.Uint256]bool)
	for _, tx := range b.Transactions {
		txHash := tx.Hash()
		if mask[txHash] {
			return errors.New("duplicated transaction in block")
		}
		mask[txHash] = true
	}

	var root common.Uint256
	copy(root[:], b.Header.TransactionsRoot[:])
	b.rebuildMerkleRoot()
	if b.Header.TransactionsRoot != root {
		return fmt.Errorf("mismatched transaction root %x and %x", b.Header.TransactionsRoot.ToArray(), root.ToArray())
	}

	return nil
}

func (b *Block) MerkleProve(i int) ([]byte, error) {
	return b.merkleTree.MerkleInclusionLeafPath(b.Transactions[i].Payload.RawData(), uint64(i), uint64(len(b.Transactions)))
}

func (b *Block) ToArray() ([]byte, error) {
	sink := common.NewZeroCopySink(nil)
	err := b.Serialization(sink)
	if err != nil {
		return nil, err
	}
	return sink.Bytes(), nil
}

func (b *Block) Hash() common.Uint256 {
	hash := b.Header.Hash()
	return *hash
}

func (b *Block) HashString() string {
	hash := b.Header.Hash()
	return hash.ToHexString()
}

func (b *Block) rebuildMerkleRoot() {
	txs := b.Transactions
	if len(txs) == 0 {
		b.Header.TransactionsRoot = common.Uint256{}
		return
	}
	tree := merkle.NewTree(0, nil, merkle.NewMemHashStore())
	for _, tx := range txs {
		tree.Append(tx.Payload.RawData())
	}
	b.merkleTree = tree
	b.Header.TransactionsRoot = tree.Root()
}
