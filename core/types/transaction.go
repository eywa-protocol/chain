package types

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
)

type transaction struct {
	Payload payload.Payload

	hash *common.Uint256
}

func (tx transaction) ToArray() []byte {
	sink := new(common.ZeroCopySink)
	err := tx.Serialization(sink)
	if err != nil {
		panic(err)
	}
	return sink.Bytes()
}

func (tx transaction) TxType() payload.TransactionType {
	return tx.Payload.TxType()
}

func (tx transaction) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteByte(byte(tx.Payload.TxType()))
	return tx.Payload.Serialization(sink)
}

func (tx *transaction) Deserialization(source *common.ZeroCopySource) error {
	txType, eof := source.NextByte()
	if eof {
		return errors.New("read tx type eof")
	}

	switch payload.TransactionType(txType) {
	case payload.InvokeType:
		var parsed payload.InvokeCode
		err := parsed.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = &parsed

	case payload.EpochType:
		var parsed payload.EpochEvent
		err := parsed.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = &parsed

	case payload.BridgeEventType:
		var parsed payload.BridgeEvent
		err := parsed.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = &parsed

	case payload.BridgeEventSolanaType:
		var parsed payload.BridgeSolanaEvent
		err := parsed.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = &parsed

	case payload.SolanaToEVMEventType:
		var parsed payload.SolanaToEVMEvent
		err := parsed.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = &parsed

	case payload.ReceiveRequestEventType:
		var parsed payload.ReceiveRequestEvent
		err := parsed.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = &parsed

	case payload.SolReceiveRequestEventType:
		var parsed payload.SolReceiveRequestEvent
		err := parsed.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = &parsed

	default:
		return fmt.Errorf("failed to unmarshal unknown tx type %d", txType)
	}
	return nil
}

func (tx *transaction) Hash() common.Uint256 {
	if tx.hash != nil {
		return *tx.hash
	}

	data := tx.Payload.RawData()
	hash := common.Uint256(sha256.Sum256(data))
	tx.hash = &hash
	return hash
}

func ToTransaction(payload payload.Payload) transaction {
	return transaction{Payload: payload}
}

func TransactionDeserialization(source *common.ZeroCopySource) (transaction, error) {
	var tx transaction
	err := tx.Deserialization(source)
	return tx, err
}

type Transactions []transaction

func (txs Transactions) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint32(uint32(len(txs)))
	for _, tx := range txs {
		if err := tx.Serialization(sink); err != nil {
			return err
		}
	}
	return nil
}

func (txs *Transactions) Deserialization(source *common.ZeroCopySource) error {
	l, eof := source.NextUint32()
	if eof {
		return errors.New("read tx length eof")
	}

	*txs = make(Transactions, 0, l)
	for i := uint32(0); i < l; i++ {
		var tx transaction
		err := tx.Deserialization(source)
		if err != nil {
			return err
		}
		*txs = append(*txs, tx)
	}
	return nil
}
