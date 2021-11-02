/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/
package states

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"testing"

	"bytes"

	"github.com/stretchr/testify/assert"
)

func TestValidatorState_Deserialize_Serialize(t *testing.T) {
	_, pubKey := bls.GenerateRandomKey()

	vs := ValidatorState{
		StateBase: StateBase{(byte)(1)},
		PublicKey: pubKey,
	}

	buf := bytes.NewBuffer(nil)
	vs.Serialize(buf)
	bs := buf.Bytes()

	var vs2 ValidatorState
	vs2.Deserialize(buf)
	assert.Equal(t, vs, vs2)

	buf = bytes.NewBuffer(bs[:len(bs)-1])
	err := vs2.Deserialize(buf)
	assert.NotNil(t, err)
}
