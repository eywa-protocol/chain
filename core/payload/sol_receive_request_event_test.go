package payload

import (
	"fmt"
	"testing"

	"github.com/eywa-protocol/chain/common"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
)

func TestSolReceiveRequestEvent_Serialization(t *testing.T) {
	bEvt := NewSolReceiveRequestEvent(&bridge.BridgeReceiveEvent{
		ReceiveRequest: bridge.ReceiveRequest{
			RequestId:   solana.PublicKey{1, 2, 3, 4, 5},
			ReceiveSide: solana.PublicKey{},
			BridgeFrom:  common.Address{},
		},
		Signature: solana.Signature{42},
		Slot:      2,
	})

	sink := common.NewZeroCopySink(nil)
	err := bEvt.Serialization(sink)
	assert.NoError(t, err)
	// t.Log(sink.Bytes())

	var bridgeEvent2 SolReceiveRequestEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bEvt, &bridgeEvent2)

	// test ToJson
	jbExpected := `{"ReqId":"0102030405000000000000000000000000000000000000000000000000000000","ReceiveSide":"0000000000000000000000000000000000000000000000000000000000000000","BridgeFrom":"0x0000000000000000000000000000000000000000","Signature":"2a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","Slot":2}`
	jb, err := bEvt.ToJson()
	fmt.Println(string(jb))
	assert.NoError(t, err)
	assert.Equal(t, jbExpected, string(jb))

	// test DestChainId
	uChainId, fromHead := bridgeEvent2.DstChainId()
	assert.Equal(t, true, fromHead)
	assert.Equal(t, uint64(0), uChainId)
}
