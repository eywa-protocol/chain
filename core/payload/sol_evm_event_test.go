package payload

import (
	"testing"

	"github.com/eywa-protocol/chain/common"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
)

func TestBridgeSolToEvmEvent_Serialize(t *testing.T) {
	bEvt := SolanaToEVMEvent{
		OriginData: bridge.BridgeEvent{
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
	}

	sink := common.NewZeroCopySink(nil)
	err := bEvt.Serialization(sink)
	assert.NoError(t, err)
	// t.Log(sink.Bytes())

	var bridgeEvent2 SolanaToEVMEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bEvt, bridgeEvent2)

	// test ToJson
	jbExpected := `{"request_type":"test","bridge_pub_key":"11111111111111111111111111111111","request_id":"11111111111111111111111111111111","selector":"dGVzdHNlbGVjdG9y","receive_side":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"opposite_bridge":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"chain_id":3,"signature":"1111111111111111111111111111111111111111111111111111111111111111","slot":3}`
	jb, err := bEvt.ToJson()
	assert.NoError(t, err)
	assert.Equal(t, jbExpected, string(jb))

	// test DestChainId
	uChainId, fromHead := bridgeEvent2.DstChainId()
	assert.Equal(t, false, fromHead)
	assert.Equal(t, uint64(3), uChainId)
}
