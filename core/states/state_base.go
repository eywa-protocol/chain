package states

import (
	"fmt"
	"io"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/serialization"
)

type StateBase struct {
	StateVersion byte
}

func (this *StateBase) Serialize(w io.Writer) error {
	serialization.WriteByte(w, this.StateVersion)
	return nil
}

func (this *StateBase) Deserialize(r io.Reader) error {
	stateVersion, err := serialization.ReadByte(r)
	if err != nil {
		return fmt.Errorf("[StateBase], StateBase Deserialize failed,%s", err)
	}
	this.StateVersion = stateVersion
	return nil
}
