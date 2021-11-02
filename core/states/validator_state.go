package states

import (
	"fmt"
	"github.com/eywa-protocol/bls-crypto/bls"
	"io"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/serialization"
)

type ValidatorState struct {
	StateBase
	PublicKey bls.PublicKey
}

func (this *ValidatorState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	buf := this.PublicKey.Marshal()
	if err := serialization.WriteVarBytes(w, buf); err != nil {
		return err
	}
	return nil
}

func (this *ValidatorState) Deserialize(r io.Reader) error {
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return fmt.Errorf("[ValidatorState], StateBase Deserialize failed, error:%s", err)
	}
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[ValidatorState], PublicKey Deserialize failed, error:%s", err)
	}
	pk, err := bls.UnmarshalPublicKey(buf)
	if err != nil {
		return fmt.Errorf("[ValidatorState], PublicKey Deserialize failed, error:%s", err)
	}
	this.PublicKey = pk
	return nil
}
