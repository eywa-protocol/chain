/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/

package types

import (
	"errors"
	"fmt"
	"io"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
)

type Block struct {
	Header       *Header
	Transactions []*Transaction
}

func (b *Block) Serialization(sink *common.ZeroCopySink) error {
	err := b.Header.Serialization(sink)
	if err != nil {
		return err
	}

	sink.WriteUint32(uint32(len(b.Transactions)))
	for _, transaction := range b.Transactions {
		err := transaction.Serialization(sink)
		if err != nil {
			return err
		}
	}
	return nil
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

	length, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}

	var hashes []common.Uint256
	mask := make(map[common.Uint256]bool)
	for i := uint32(0); i < length; i++ {
		transaction := new(Transaction)
		// note currently all transaction in the block shared the same source
		err := transaction.Deserialization(source)
		if err != nil {
			return err
		}
		txhash := transaction.Hash()
		if mask[txhash] {
			return errors.New("duplicated transaction in block")
		}
		mask[txhash] = true
		hashes = append(hashes, txhash)
		self.Transactions = append(self.Transactions, transaction)
	}
	root := common.ComputeMerkleRoot(hashes)
	if self.Header.TransactionsRoot != root {
		return fmt.Errorf("mismatched transaction root %x and %x", self.Header.TransactionsRoot.ToArray(), root.ToArray())
	}

	return nil
}

func (b *Block) ToArray() []byte {
	sink := common.NewZeroCopySink(nil)
	b.Serialization(sink)
	return sink.Bytes()
}

func (b *Block) Hash() common.Uint256 {
	return b.Header.Hash()
}

func (b *Block) Type() common.InventoryType {
	return common.BLOCK
}

func (b *Block) RebuildMerkleRoot() {
	txs := b.Transactions
	hashes := make([]common.Uint256, 0, len(txs))
	for _, tx := range txs {
		hashes = append(hashes, tx.Hash())
	}
	hash := common.ComputeMerkleRoot(hashes)
	b.Header.TransactionsRoot = hash
}
