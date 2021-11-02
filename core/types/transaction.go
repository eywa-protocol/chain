package types

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/eywa-protocol/bls-crypto/bls"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/payload"
	"sort"
)

const MAX_TX_SIZE = 1024 * 1024 // The max size of a transaction to prevent DOS attacks

type CoinType byte

const (
	ONG CoinType = iota
	ETH
	BTC
)

type Transaction struct {
	TxType     TransactionType
	Nonce      uint32
	ChainID    uint64
	GasLimit   uint64
	GasPrice   uint64
	Payload    Payload
	Attributes []byte //this must be 0 now, Attribute Array length use VarUint encoding, so byte is enough for extension
	Payer      common.Address
	CoinType   CoinType
	Sig        Sig

	Raw        []byte // raw transaction data
	hash       common.Uint256
	SignedAddr []common.Address // this is assigned when passed signature verification
}

func (tx *Transaction) SerializeUnsigned(sink *common.ZeroCopySink) error {
	sink.WriteByte(byte(tx.TxType))
	sink.WriteUint32(tx.Nonce)
	sink.WriteUint64(tx.ChainID)
	sink.WriteUint64(tx.GasLimit)
	sink.WriteUint64(tx.GasPrice)
	//Payload
	if tx.Payload == nil {
		return errors.New("transaction payload is nil")
	}
	switch pl := tx.Payload.(type) {
	case *payload.InvokeCode:
		pl.Serialization(sink)
	case *payload.Epoch:
		pl.Serialization(sink)
	default:
		return errors.New("wrong transaction payload type")
	}
	if len(tx.Attributes) > MAX_ATTRIBUTES_LEN {
		return fmt.Errorf("attributes length %d over max length %d", tx.Attributes, MAX_ATTRIBUTES_LEN)
	}
	sink.WriteVarBytes(tx.Attributes)
	sink.WriteAddress(tx.Payer)
	sink.WriteByte(byte(tx.CoinType))
	return nil
}

// Serialize the Transaction
func (tx *Transaction) Serialization(sink *common.ZeroCopySink) error {
	if err := tx.SerializeUnsigned(sink); err != nil {
		return err
	}

	if err := tx.Sig.Serialize(sink); err != nil {
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

	var sig Sig
	if err := sig.Deserialize(source); err != nil {
		return err
	}
	tx.Sig = sig

	pend := source.Pos()
	lenAll := pend - pstart
	if lenAll > MAX_TX_SIZE {
		return fmt.Errorf("execced max transaction size:%d", lenAll)
	}
	source.BackUp(lenAll)
	tx.Raw, _ = source.NextBytes(lenAll)
	return nil
}

func (tx *Transaction) DeserializationUnsigned(source *common.ZeroCopySource) error {
	var eof bool
	txType, eof := source.NextByte()
	if eof {
		return errors.New("[deserializationUnsigned] read txType error")
	}
	tx.TxType = TransactionType(txType)
	tx.Nonce, eof = source.NextUint32()
	if eof {
		return errors.New("[deserializationUnsigned] read nonce error")
	}
	tx.ChainID, eof = source.NextUint64()
	if eof {
		return errors.New("[deserializationUnsigned] read chainid error")
	}
	tx.GasLimit, eof = source.NextUint64()
	if eof {
		return errors.New("[deserializationUnsigned] read gaslimit error")
	}
	tx.GasPrice, eof = source.NextUint64()
	if eof {
		return errors.New("[deserializationUnsigned] read gasprice error")
	}

	switch tx.TxType {
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

	default:
		return fmt.Errorf("unsupported tx type %v", tx.Type())
	}
	tx.Attributes, eof = source.NextVarBytes()
	if eof {
		return errors.New("[deserializationUnsigned] read attributes error")
	}
	if len(tx.Attributes) > MAX_ATTRIBUTES_LEN {
		return fmt.Errorf("[deserializationUnsigned] attributes length %d over max limit %d", tx.Attributes, MAX_ATTRIBUTES_LEN)
	}
	tx.Payer, eof = source.NextAddress()
	if eof {
		return errors.New("[deserializationUnsigned] read payer error")
	}
	coinType, eof := source.NextByte()
	if eof {
		return errors.New("[deserializationUnsigned] read coinType error")
	}
	tx.CoinType = CoinType(coinType)
	if tx.CoinType != ONG {
		return errors.New("[deserializationUnsigned] unsupported coinType")
	}
	return nil
}

type Sig struct {
	SigData bls.Signature // aggregated signature of all who signed
	PubKey  bls.PublicKey // aggregated public key of all who signed
	M       uint64        // bitmask of all who signed
}

func (this *Sig) Serialize(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes(this.SigData.Marshal())
	sink.WriteVarBytes(this.PubKey.Marshal())
	sink.WriteUint64(this.M)
	return nil
}

func (this *Sig) Deserialize(source *common.ZeroCopySource) error {
	data, eof := source.NextVarBytes()
	if eof {
		return errors.New("[Sig] deserialize read sigData error")
	}
	sig, err := bls.UnmarshalSignature(data)
	if err != nil {
		return err
	}
	this.SigData = sig

	data, eof = source.NextVarBytes()
	if eof {
		return errors.New("[Sig] deserialize read pubKey error")
	}
	pk, err := bls.UnmarshalPublicKey(data)
	if err != nil {
		return err
	}
	this.PubKey = pk

	m, eof := source.NextUint64()
	if eof {
		return errors.New("[Sig] deserialize read M error")
	}
	this.M = m
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
	Invoke            TransactionType = 0xd1
	AddNode           TransactionType = 0xd2
	AddCrosschainCall TransactionType = 0xd3
	Epoch             TransactionType = 0xa1
	AddUpTime         TransactionType = 0xd5
)

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

const MULTI_SIG_MAX_PUBKEY_SIZE = 16

func EncodeMultiPubKeyProgramInto(sink *common.ZeroCopySink, pubkeys []bls.PublicKey, m uint16) error {
	n := len(pubkeys)
	if !(1 <= m && int(m) <= n && n > 1 && n <= MULTI_SIG_MAX_PUBKEY_SIZE) {
		return errors.New("wrong multi-sig param")
	}
	pubkeys = SortPublicKeys(pubkeys)
	sink.WriteUint16(uint16(len(pubkeys)))
	for _, pubkey := range pubkeys {
		//fmt.Printf("\npubkey %v", common.ToHexString(pubkey.Marshal()))
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
