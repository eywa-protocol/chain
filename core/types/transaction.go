package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/sirupsen/logrus"
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
	len, eof := source.NextUint32()
	if eof {
		return errors.New("read tx length eof")
	}

	*txs = make(Transactions, 0, len)
	for i := uint32(0); i < len; i++ {
		var tx transaction
		err := tx.Deserialization(source)
		if err != nil {
			return err
		}
		*txs = append(*txs, tx)
	}
	return nil
}

func (txs Transactions) LogFields() logrus.Fields {
	reqIds := make([]string, 0, len(txs))
	txIds := make([]string, 0, len(txs))
	for _, tx := range txs {
		id := tx.Payload.RequestId()
		reqIds = append(reqIds, hex.EncodeToString(id[:]))
		txIds = append(txIds, hex.EncodeToString(tx.Payload.SrcTxHash()))
	}
	return logrus.Fields{
		"req_ids": reqIds,
		"tx_ids":  txIds,
	}
}
