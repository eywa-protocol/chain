package payload

import (
	"testing"

	"github.com/eywa-protocol/chain/common"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
)

func TestBridgeSolToEvmEvent_Serialize(t *testing.T) {
	bEvt := NewSolanaToEVMEvent(&bridge.BridgeEvent{
		OracleRequest: bridge.OracleRequest{
			RequestType:    "test",
			BridgePubKey:   solana.PublicKey{},
			RequestId:      solana.PublicKey{},
			Selector:       []byte("testselector"),
			ReceiveSide:    common.Address{},
			OppositeBridge: common.Address{},
			ChainId:        uint64(3),
		},
		Signature: solana.Signature{},
		Slot:      uint64(3),
	},
	)

	sink := common.NewZeroCopySink(nil)
	err := bEvt.Serialization(sink)
	assert.NoError(t, err)
	// t.Log(sink.Bytes())

	var bridgeEvent2 SolanaToEVMEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bEvt.Data(), bridgeEvent2.Data())

	// test ToJson
	jbExpected := `{"RequestType":"test","BridgePubKey":"0000000000000000000000000000000000000000000000000000000000000000","ReqId":"0000000000000000000000000000000000000000000000000000000000000000","Selector":"dGVzdHNlbGVjdG9y","ReceiveSide":"0x0000000000000000000000000000000000000000","OppositeBridge":"0x0000000000000000000000000000000000000000","ChainId":3,"Signature":"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","Slot":3}`
	jb, err := bEvt.ToJson()
	assert.NoError(t, err)
	assert.Equal(t, jbExpected, string(jb))

	// test DestChainId
	uChainId, fromHead := bridgeEvent2.DstChainId()
	assert.Equal(t, false, fromHead)
	assert.Equal(t, uint64(3), uChainId)
}
