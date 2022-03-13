package types

import (
	"errors"
	"fmt"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
)

type Block struct {
	Header       *Header
	Transactions Transactions
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
	block := &Block{
		Header:       header,
		Transactions: transactions,
	}
	block.rebuildMerkleRoot()
	return block
}

// if no error, ownership of param raw is transfered to Transaction
func BlockFromRawBytes(raw []byte) (*Block, error) {
	source := common.NewZeroCopySource(raw)
	block := &Block{}
	err := block.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (self *Block) Deserialization(source *common.ZeroCopySource) error {
	if self.Header == nil {
		self.Header = new(Header)
	}
	err := self.Header.Deserialization(source)
	if err != nil {
		return err
	}

	return self.Transactions.Deserialization(source)
}

func (self *Block) VerifyIntegrity() error {
	hashes := make([]common.Uint256, 0, len(self.Transactions))
	mask := make(map[common.Uint256]bool)
	for _, tx := range self.Transactions {
		txhash := tx.Hash()
		if mask[txhash] {
			return errors.New("duplicated transaction in block")
		}
		mask[txhash] = true
		hashes = append(hashes, txhash)
	}

	root := CalculateMerkleRoot(hashes)
	if self.Header.TransactionsRoot != root {
		return fmt.Errorf("mismatched transaction root %x and %x", self.Header.TransactionsRoot.ToArray(), root.ToArray())
	}

	return nil
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
	return b.Header.Hash()
}

func (b *Block) HashString() string {
	hash := b.Header.Hash()
	return hash.ToHexString()
}

func (b *Block) rebuildMerkleRoot() {
	txs := b.Transactions
	hashes := make([]common.Uint256, 0, len(txs))
	for _, tx := range txs {
		hashes = append(hashes, tx.Hash())
	}
	hash := CalculateMerkleRoot(hashes)
	b.Header.TransactionsRoot = hash
}
