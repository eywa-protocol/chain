package payload

import (
	"fmt"
	"github.com/eywa-protocol/chain/common"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
	"testing"
)

func TestSolReceiveRequestEvent_Serialization(t *testing.T) {
	bEvt := SolReceiveRequestEvent{
		OriginData: bridge.BridgeReceiveEvent{
			ReceiveRequest: bridge.ReceiveRequest{
				RequestId:   solana.PublicKey{},
				ReceiveSide: solana.PublicKey{},
				BridgeFrom:  common.Address{},
			},
			Signature: solana.Signature{},
			Slot:      2,
		},
	}

	sink := common.NewZeroCopySink(nil)
	err := bEvt.Serialization(sink)
	assert.NoError(t, err)
	// t.Log(sink.Bytes())

	var bridgeEvent2 SolReceiveRequestEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bEvt, bridgeEvent2)

	// test ToJson
	jbExpected := `{"req_id":"11111111111111111111111111111111","receive_side":"11111111111111111111111111111111","bridge_from":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"signature":"1111111111111111111111111111111111111111111111111111111111111111","slot":2}`
	jb, err := bEvt.ToJson()
	fmt.Println(string(jb))
	assert.NoError(t, err)
	assert.Equal(t, jbExpected, string(jb))

	// test DestChainId
	uChainId, fromHead := bridgeEvent2.DstChainId()
	assert.Equal(t, true, fromHead)
	assert.Equal(t, uint64(0), uChainId)
}
