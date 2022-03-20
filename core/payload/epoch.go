package payload

import (
	"encoding/binary"
	"fmt"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
)

type Epoch struct {
	Number         uint32          // Number of this epoch
	EpochPublicKey bls.PublicKey   // Aggregated public key of all participants of the current epoch
	PublicKeys     []bls.PublicKey // Public keys of all nodes
}

func (tx *Epoch) TxType() TransactionType {
	return EpochType
}

func (self *Epoch) Deserialization(source *common.ZeroCopySource) error {
	number, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("Epoch.Number deserialize eof")
	}
	epochPublicKeyRaw, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Epoch.EpochPublicKey deserialize eof")
	}
	self.Number = number

	epochPublicKey, err := bls.UnmarshalPublicKey(epochPublicKeyRaw)
	if err != nil {
		return fmt.Errorf("Epoch.EpochPublicKey deserialize error %v", err)
	}
	self.EpochPublicKey = epochPublicKey

	length, eof := source.NextUint8()
	if eof {
		return fmt.Errorf("Epoch.len(PublicKeys) deserialize eof")
	}

	self.PublicKeys = make([]bls.PublicKey, 0, length)
	for i := uint8(0); i < length; i++ {
		publicKeyRaw, eof := source.NextVarBytes()
		if eof {
			return fmt.Errorf("Epoch.PublicKey deserialize eof")
		}
		publicKey, err := bls.UnmarshalPublicKey(publicKeyRaw)
		if err != nil {
			return fmt.Errorf("Epoch.PublicKey[%d] deserialize error %v", i, err)
		}
		self.PublicKeys = append(self.PublicKeys, publicKey)
	}

	return nil
}

func (self *Epoch) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint32(self.Number)
	sink.WriteVarBytes(self.EpochPublicKey.Marshal())
	sink.WriteUint8(uint8(len(self.PublicKeys)))
	for _, key := range self.PublicKeys {
		sink.WriteVarBytes(key.Marshal())
	}
	return nil
}

func (self *Epoch) RawData() []byte {
	epochNumRaw := make([]byte, 4)
	binary.BigEndian.PutUint32(epochNumRaw, self.Number)

	sink := common.NewZeroCopySink(nil)
	sink.WriteBytes(epochNumRaw)
	sink.WriteVarBytes(self.EpochPublicKey.Marshal())
	sink.WriteUint8(uint8(len(self.PublicKeys)))
	return sink.Bytes()
}
