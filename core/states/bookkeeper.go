package states

import (
	"bytes"
	"github.com/eywa-protocol/bls-crypto/bls"
	"io"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/serialization"
)

type BookkeeperState struct {
	StateBase
	CurrBookkeeper []bls.PublicKey
	NextBookkeeper []bls.PublicKey
}

func (this *BookkeeperState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteUint32(w, uint32(len(this.CurrBookkeeper)))
	for _, v := range this.CurrBookkeeper {
		buf := v.Marshal()
		err := serialization.WriteVarBytes(w, buf)
		if err != nil {
			return err
		}
	}
	serialization.WriteUint32(w, uint32(len(this.NextBookkeeper)))
	for _, v := range this.NextBookkeeper {

		buf := v.Marshal()
		err := serialization.WriteVarBytes(w, buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *BookkeeperState) Deserialize(r io.Reader) error {
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
		this.CurrBookkeeper = append(this.CurrBookkeeper, key)
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
		this.NextBookkeeper = append(this.NextBookkeeper, key)
	}
	return nil
}

func (v *BookkeeperState) ToArray() []byte {
	b := new(bytes.Buffer)
	v.Serialize(b)
	return b.Bytes()
}
