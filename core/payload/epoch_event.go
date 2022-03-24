package payload

import (
	"encoding/binary"
	"fmt"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
)

type EpochEvent struct {
	Number         uint32          // Number of this epoch
	EpochPublicKey bls.PublicKey   // Aggregated public key of all participants of the current epoch
	SourceTx       common.Uint256  // Governance blockchain transaction that caused this epoch change
	PublicKeys     []bls.PublicKey // Public keys of all nodes (informational, not included in hashing)
}

func NewEpochEvent(num uint32, pub bls.PublicKey, tx common.Uint256, keys []bls.PublicKey) *EpochEvent {
	return &EpochEvent{
		Number:         num,
		EpochPublicKey: pub,
		SourceTx:       tx,
		PublicKeys:     keys,
	}
}

func (e *EpochEvent) TxType() TransactionType {
	return EpochType
}

func (e *EpochEvent) Deserialization(source *common.ZeroCopySource) error {
	number, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("Epoch.Number deserialize eof")
	}
	e.Number = number

	epochPublicKeyRaw, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Epoch.EpochPublicKey deserialize eof")
	}
	epochPublicKey, err := bls.UnmarshalPublicKey(epochPublicKeyRaw)
	if err != nil {
		return fmt.Errorf("Epoch.EpochPublicKey deserialize error %v", err)
	}
	e.EpochPublicKey = epochPublicKey

	sourceTx, eof := source.NextBytes(common.UINT256_SIZE)
	if eof {
		return fmt.Errorf("Epoch.SourceTx deserialize eof")
	}
	e.SourceTx, _ = common.Uint256ParseFromBytes(sourceTx)

	length, eof := source.NextUint8()
	if eof {
		return fmt.Errorf("Epoch.len(PublicKeys) deserialize eof")
	}

	e.PublicKeys = make([]bls.PublicKey, 0, length)
	for i := uint8(0); i < length; i++ {
		publicKeyRaw, eof := source.NextVarBytes()
		if eof {
			return fmt.Errorf("Epoch.PublicKey deserialize eof")
		}
		publicKey, err := bls.UnmarshalPublicKey(publicKeyRaw)
		if err != nil {
			return fmt.Errorf("Epoch.PublicKey[%d] deserialize error %v", i, err)
		}
		e.PublicKeys = append(e.PublicKeys, publicKey)
	}

	return nil
}

func (e *EpochEvent) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint32(e.Number)
	sink.WriteVarBytes(e.EpochPublicKey.Marshal())
	sink.WriteBytes(e.SourceTx[:])
	sink.WriteUint8(uint8(len(e.PublicKeys)))
	for _, key := range e.PublicKeys {
		sink.WriteVarBytes(key.Marshal())
	}
	return nil
}

func (e *EpochEvent) RawData() []byte {
	epochNumRaw := make([]byte, 4)
	binary.BigEndian.PutUint32(epochNumRaw, e.Number)

	sink := common.NewZeroCopySink(nil)
	sink.WriteBytes(epochNumRaw)
	sink.WriteVarBytes(e.EpochPublicKey.Marshal())
	sink.WriteBytes(e.SourceTx[:])
	sink.WriteUint8(uint8(len(e.PublicKeys)))
	return sink.Bytes()
}
