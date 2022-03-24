package payload

import (
	"testing"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
	"github.com/stretchr/testify/assert"
)

func TestEpochEvent_Serialize(t *testing.T) {
	epoch, err := bls.ReadPublicKey("1d65becbb891b6e69951febbc4ac066343670b34d84777a077c06871beb9c07f28be8a6fa825e9d615f56f0dbcd728b46e42b4ae2a611e2ab919a1de923ae7ed0f1c89b508af036f52c2215a04e13a7a5e891d9220d3d8751dc0525b81fca3051dc2e58a167c412941bd1adeb29f5a0beb5d26e748e8ca55e508deadead1ea5e")
	assert.NoError(t, err)

	event := NewEpochEvent(123, epoch, common.UINT256_EMPTY, []bls.PublicKey{epoch, epoch, epoch})

	sink := common.NewZeroCopySink(nil)
	err = event.Serialization(sink)
	assert.NoError(t, err)
	var received EpochEvent
	err = received.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, *event, received)

	// test ToJson
	jbExpected := `{"Number":123,"EpochPublicKey":"1d65becbb891b6e69951febbc4ac066343670b34d84777a077c06871beb9c07f28be8a6fa825e9d615f56f0dbcd728b46e42b4ae2a611e2ab919a1de923ae7ed0f1c89b508af036f52c2215a04e13a7a5e891d9220d3d8751dc0525b81fca3051dc2e58a167c412941bd1adeb29f5a0beb5d26e748e8ca55e508deadead1ea5e","SourceTx":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"PublicKeys":["1d65becbb891b6e69951febbc4ac066343670b34d84777a077c06871beb9c07f28be8a6fa825e9d615f56f0dbcd728b46e42b4ae2a611e2ab919a1de923ae7ed0f1c89b508af036f52c2215a04e13a7a5e891d9220d3d8751dc0525b81fca3051dc2e58a167c412941bd1adeb29f5a0beb5d26e748e8ca55e508deadead1ea5e","1d65becbb891b6e69951febbc4ac066343670b34d84777a077c06871beb9c07f28be8a6fa825e9d615f56f0dbcd728b46e42b4ae2a611e2ab919a1de923ae7ed0f1c89b508af036f52c2215a04e13a7a5e891d9220d3d8751dc0525b81fca3051dc2e58a167c412941bd1adeb29f5a0beb5d26e748e8ca55e508deadead1ea5e","1d65becbb891b6e69951febbc4ac066343670b34d84777a077c06871beb9c07f28be8a6fa825e9d615f56f0dbcd728b46e42b4ae2a611e2ab919a1de923ae7ed0f1c89b508af036f52c2215a04e13a7a5e891d9220d3d8751dc0525b81fca3051dc2e58a167c412941bd1adeb29f5a0beb5d26e748e8ca55e508deadead1ea5e"]}`
	jb, err := received.ToJson()
	assert.NoError(t, err)
	assert.Equal(t, jbExpected, string(jb))

	// test DestChainId
	uChainId, fromHead := received.DstChainId()
	assert.Equal(t, true, fromHead)
	assert.Equal(t, uint64(0), uChainId)

}
