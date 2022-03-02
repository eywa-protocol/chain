package types

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
)

var (
	ErrNotSupportedTxType = errors.New("not supported tx type")
)

const MAX_TX_SIZE = 1024 * 1024 // The max size of a transaction to prevent DOS attacks

type Transaction struct {
	TxType  TransactionType
	Payload Payload

	hash common.Uint256
}

func (tx *Transaction) SerializeUnsigned(sink *common.ZeroCopySink) error {
	sink.WriteByte(byte(tx.TxType))

	if tx.Payload == nil {
		return errors.New("transaction payload is nil")
	}
	switch pl := tx.Payload.(type) {
	case *payload.InvokeCode:
		pl.Serialization(sink)
	case *payload.Epoch:
		pl.Serialization(sink)
	case *payload.BridgeEvent:
		pl.Serialization(sink)
	case *payload.BridgeSolanaEvent:
		pl.Serialization(sink)
	case *payload.SolanaToEVMEvent:
		pl.Serialization(sink)
	case *payload.ReceiveRequestEvent:
		pl.Serialization(sink)
	default:
		return errors.New("wrong transaction payload type")
	}
	return nil
}

// Serialize the Transaction
func (tx *Transaction) Serialization(sink *common.ZeroCopySink) error {
	if err := tx.SerializeUnsigned(sink); err != nil {
		return err
	}

	return nil
}

// if no error, ownership of param raw is transfered to Transaction
func TransactionFromRawBytes(raw []byte) (*Transaction, error) {
	if len(raw) > MAX_TX_SIZE {
		return nil, errors.New("execced max transaction size")
	}
	source := common.NewZeroCopySource(raw)
	tx := &Transaction{}
	err := tx.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Transaction has internal reference of param `source`
func (tx *Transaction) Deserialization(source *common.ZeroCopySource) error {
	pstart := source.Pos()
	if err := tx.DeserializationUnsigned(source); err != nil {
		return err
	}
	pos := source.Pos()
	lenUnsigned := pos - pstart
	source.BackUp(lenUnsigned)
	rawUnsigned, eof := source.NextBytes(lenUnsigned)
	if eof {
		return fmt.Errorf("read unsigned code error")
	}
	temp := sha256.Sum256(rawUnsigned)
	tx.hash = sha256.Sum256(temp[:])

	return nil
}

func (tx *Transaction) DeserializationUnsigned(source *common.ZeroCopySource) error {
	var eof bool
	txType, eof := source.NextByte()
	if eof {
		return errors.New("[deserializationUnsigned] read txType error")
	}
	tx.TxType = TransactionType(txType)

	switch tx.TxType {

	case ReceiveRequestEvent:
		pl := new(payload.ReceiveRequestEvent)
		err := pl.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = pl

	case BridgeEvent:
		pl := new(payload.InvokeCode)
		err := pl.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = pl

	case Invoke:
		pl := new(payload.InvokeCode)
		err := pl.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = pl

	case Epoch:
		pl := new(payload.Epoch)
		err := pl.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = pl

	case BridgeEventSolana:
		pl := new(payload.InvokeCode)
		err := pl.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = pl
	case SolanaToEVMEvent:
		pl := new(payload.SolanaToEVMEvent)
		err := pl.Deserialization(source)
		if err != nil {
			return err
		}
		tx.Payload = pl
	default:
		return fmt.Errorf("unsupported tx type %v", tx.Type())
	}
	return nil
}

/*func (self *Transaction) GetSignatureAddresses() ([]common.Address, error) {
	if len(self.SignedAddr) == 0 {
		addrs := make([]common.Address, 0, len(self.Sigs))
		for _, prog := range self.Sigs {
			if len(prog.PubKeys) == 0 {
				return nil, errors.New("[GetSignatureAddresses] no public key")
			} else if len(prog.PubKeys) == 1 {
				buf := prog.PubKeys[0].Marshal()
				addrs = append(addrs, common.AddressFromVmCode(buf))
			} else {
				sink := common.NewZeroCopySink(nil)
				if err := EncodeMultiPubKeyProgramInto(sink, prog.PubKeys, prog.M); err != nil {
					return nil, err
				}
				addrs = append(addrs, common.AddressFromVmCode(sink.Bytes()))
			}
		}
		self.SignedAddr = addrs
	}
	return self.SignedAddr, nil
}*/

type TransactionType byte

const (
	Invoke              TransactionType = 0xd1
	Node                TransactionType = 0xd2
	Epoch               TransactionType = 0x22
	UpTime              TransactionType = 0xd4
	BridgeEvent         TransactionType = 0x1f
	BridgeEventSolana   TransactionType = 0x20
	SolanaToEVMEvent    TransactionType = 0x21
	ReceiveRequestEvent TransactionType = 0x23
)

func (tt TransactionType) String() string {
	switch tt {
	case Invoke:
		return "invoke"
	case Node:
		return "node"
	case Epoch:
		return "epoch"
	case UpTime:
		return "up_time"
	case BridgeEvent:
		return "bridge_event"
	case ReceiveRequestEvent:
		return "receive_request_event"
	default:
		return "unknown"
	}
}

// Payload define the func for loading the payload data
// base on payload type which have different structure
type Payload interface {
	Deserialization(source *common.ZeroCopySource) error

	Serialization(sink *common.ZeroCopySink)
}

func (tx *Transaction) ToArray() []byte {
	sink := new(common.ZeroCopySink)
	tx.Serialization(sink)
	return sink.Bytes()
}

func (tx *Transaction) Hash() common.Uint256 {
	return tx.hash
}

func (tx *Transaction) Type() common.InventoryType {
	return common.TRANSACTION
}

func (tx *Transaction) LogHash() (string, error) {
	if tx.TxType == BridgeEvent {
		sink := common.NewZeroCopySink(nil)
		tx.Payload.Serialization(sink)
		var bridgeEvent payload.BridgeEvent
		if err := bridgeEvent.Deserialization(common.NewZeroCopySource(sink.Bytes())); err != nil {
			return "", err
		}
		return bridgeEvent.OriginData.Raw.TxHash.String(), nil
	}

	if tx.TxType == BridgeEventSolana {
		sink := common.NewZeroCopySink(nil)
		tx.Payload.Serialization(sink)
		var bridgeEvent payload.BridgeSolanaEvent
		if err := bridgeEvent.Deserialization(common.NewZeroCopySource(sink.Bytes())); err != nil {
			return "", err
		}
		return bridgeEvent.OriginData.Raw.TxHash.String(), nil
	}

	if tx.TxType == SolanaToEVMEvent {
		sink := common.NewZeroCopySink(nil)
		tx.Payload.Serialization(sink)
		var bridgeEvent payload.SolanaToEVMEvent
		if err := bridgeEvent.Deserialization(common.NewZeroCopySource(sink.Bytes())); err != nil {
			return "", err
		}
		return bridgeEvent.OriginData.Signature.String(), nil
	}

	if tx.TxType == ReceiveRequestEvent {
		sink := common.NewZeroCopySink(nil)
		tx.Payload.Serialization(sink)
		var bridgeEvent payload.ReceiveRequestEvent
		if err := bridgeEvent.Deserialization(common.NewZeroCopySource(sink.Bytes())); err != nil {
			return "", err
		}
		return bridgeEvent.OriginData.Raw.TxHash.String(), nil
	}

	return "", fmt.Errorf("log hash %w [%s]", ErrNotSupportedTxType, tx.TxType.String())

}

const MULTI_SIG_MAX_PUBKEY_SIZE = 16

func EncodeMultiPubKeyProgramInto(sink *common.ZeroCopySink, pubkeys []bls.PublicKey, m uint16) error {
	n := len(pubkeys)
	if !(1 <= m && int(m) <= n && n > 1 && n <= MULTI_SIG_MAX_PUBKEY_SIZE) {
		return errors.New("wrong multi-sig param")
	}
	pubkeys = SortPublicKeys(pubkeys)
	sink.WriteUint16(uint16(len(pubkeys)))
	for _, pubkey := range pubkeys {
		// fmt.Printf("\npubkey %v", common.ToHexString(pubkey.Marshal()))
		key := pubkey.Marshal()
		sink.WriteVarBytes(key)
	}
	sink.WriteUint16(m)

	return nil
}

func SortPublicKeys(list []bls.PublicKey) []bls.PublicKey {
	pl := publicKeyList(list)
	sort.Sort(pl)
	return pl
}

type publicKeyList []bls.PublicKey

func (p publicKeyList) Len() int {
	return len(p)
}

func (p publicKeyList) Less(i, j int) bool {
	return bytes.Compare(p[i].Marshal(), p[j].Marshal()) > 0
}

func (p publicKeyList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
