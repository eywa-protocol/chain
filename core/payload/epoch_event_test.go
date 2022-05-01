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

	event := NewEpochEvent(123, common.UINT256_EMPTY, []bls.PublicKey{epoch, epoch, epoch}, []string{"one", "two", "three"})

	sink := common.NewZeroCopySink(nil)
	err = event.Serialization(sink)
	assert.NoError(t, err)
	var received EpochEvent
	err = received.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, *event, received)

	// test ToJson
	jbExpected := `{"Number":123,"EpochPublicKey":"110eb22c46a82e9c3be63df6a061537f3de17d84a66fe5a491a5aced21ef0bc101a9b68af82033d3e9cc8ae964dd6ae998926dee2df8a9b323891afdc76b6956005ee00604c14856945452b6c2d055535cf3a3325ef43b44f2bb6d7e497543f1291853b24e4ebf74cfce2087b2594e54503c88d824e68e838ff85b6b8a00ded5","SourceTx":"0000000000000000000000000000000000000000000000000000000000000000","PublicKeys":["1d65becbb891b6e69951febbc4ac066343670b34d84777a077c06871beb9c07f28be8a6fa825e9d615f56f0dbcd728b46e42b4ae2a611e2ab919a1de923ae7ed0f1c89b508af036f52c2215a04e13a7a5e891d9220d3d8751dc0525b81fca3051dc2e58a167c412941bd1adeb29f5a0beb5d26e748e8ca55e508deadead1ea5e","1d65becbb891b6e69951febbc4ac066343670b34d84777a077c06871beb9c07f28be8a6fa825e9d615f56f0dbcd728b46e42b4ae2a611e2ab919a1de923ae7ed0f1c89b508af036f52c2215a04e13a7a5e891d9220d3d8751dc0525b81fca3051dc2e58a167c412941bd1adeb29f5a0beb5d26e748e8ca55e508deadead1ea5e","1d65becbb891b6e69951febbc4ac066343670b34d84777a077c06871beb9c07f28be8a6fa825e9d615f56f0dbcd728b46e42b4ae2a611e2ab919a1de923ae7ed0f1c89b508af036f52c2215a04e13a7a5e891d9220d3d8751dc0525b81fca3051dc2e58a167c412941bd1adeb29f5a0beb5d26e748e8ca55e508deadead1ea5e"],"HostIds":["one","two","three"]}`
	jb, err := received.ToJson()
	assert.NoError(t, err)
	assert.Equal(t, jbExpected, string(jb))

	// test DestChainId
	uChainId, fromHead := received.DstChainId()
	assert.Equal(t, true, fromHead)
	assert.Equal(t, uint64(0), uChainId)

}
