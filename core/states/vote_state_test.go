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

func TestVoteState_Deserialize_Serialize(t *testing.T) {
	_, pubKey1 := bls.GenerateRandomKey()
	_, pubKey2 := bls.GenerateRandomKey()

	vs := VoteState{
		StateBase:  StateBase{(byte)(1)},
		PublicKeys: []bls.PublicKey{pubKey1, pubKey2},
		Count:      10,
	}

	buf := bytes.NewBuffer(nil)
	vs.Serialize(buf)
	bs := buf.Bytes()

	var vs2 VoteState
	vs2.Deserialize(buf)
	assert.Equal(t, vs, vs2)

	buf = bytes.NewBuffer(bs[:len(bs)-1])
	err := vs2.Deserialize(buf)
	assert.NotNil(t, err)
}
