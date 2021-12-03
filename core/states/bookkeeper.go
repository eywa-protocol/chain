package states

import (
	"bytes"
	"github.com/eywa-protocol/bls-crypto/bls"
	"io"

	"github.com/eywa-protocol/chain/common/serialization"
)

type EpochState struct {
	StateBase
	CurrEpoch []bls.PublicKey
	NextEpoch []bls.PublicKey
}

func (this *EpochState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteUint32(w, uint32(len(this.CurrEpoch)))
	for _, v := range this.CurrEpoch {
		buf := v.Marshal()
		err := serialization.WriteVarBytes(w, buf)
		if err != nil {
			return err
		}
	}
	serialization.WriteUint32(w, uint32(len(this.NextEpoch)))
	for _, v := range this.NextEpoch {

		buf := v.Marshal()
		err := serialization.WriteVarBytes(w, buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *EpochState) Deserialize(r io.Reader) error {
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return err
	}
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	for i := 0; i < int(n); i++ {
		buf, err := serialization.ReadVarBytes(r)
		if err != nil {
			return err
		}
		key, err := bls.UnmarshalPublicKey(buf)
		this.CurrEpoch = append(this.CurrEpoch, key)
	}

	n, err = serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	for i := 0; i < int(n); i++ {
		buf, err := serialization.ReadVarBytes(r)
		if err != nil {
			return err
		}
		key, err := bls.UnmarshalPublicKey(buf)
		this.NextEpoch = append(this.NextEpoch, key)
	}
	return nil
}

func (v *EpochState) ToArray() []byte {
	b := new(bytes.Buffer)
	v.Serialize(b)
	return b.Bytes()
}
