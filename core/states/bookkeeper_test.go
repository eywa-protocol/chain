package states

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"testing"

	"bytes"

	"github.com/stretchr/testify/assert"
)

func TestEpoch_Deserialize_Serialize(t *testing.T) {
	_, pubKey1 := bls.GenerateRandomKey()
	_, pubKey2 := bls.GenerateRandomKey()
	_, pubKey3 := bls.GenerateRandomKey()
	_, pubKey4 := bls.GenerateRandomKey()

	bk := EpochState{
		StateBase: StateBase{(byte)(1)},
		CurrEpoch: []bls.PublicKey{pubKey1, pubKey2},
		NextEpoch: []bls.PublicKey{pubKey3, pubKey4},
	}

	buf := bytes.NewBuffer(nil)
	bk.Serialize(buf)
	bs := buf.Bytes()

	var bk2 EpochState
	bk2.Deserialize(buf)
	assert.Equal(t, bk, bk2)

	buf = bytes.NewBuffer(bs[:len(bs)-1])
	err := bk2.Deserialize(buf)
	assert.NotNil(t, err)
}
