package payload

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
)

func TestBridgeSolanaEvent_Serialize(t *testing.T) {
	bridgeEvent := BridgeSolanaEvent{
		OriginData: wrappers.BridgeOracleRequestSolana{
			RequestType: "setRequest",
			Bridge:      [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 90, 1, 2, 3, 4, 5, 6, 7, 78, 9, 0, 1, 2, 2, 3, 43, 4, 4, 5, 5, 56, 23},
			ChainId:     big.NewInt(94),
		}}

	sink := common.NewZeroCopySink(nil)
	err := bridgeEvent.Serialization(sink)
	assert.NoError(t, err)
	var bridgeEvent2 BridgeSolanaEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bridgeEvent, bridgeEvent2)

	// test ToJson
	jbExpected := `{"RequestType":"setRequest","Bridge":[1,2,3,4,5,6,7,8,90,1,2,3,4,5,6,7,78,9,0,1,2,2,3,43,4,4,5,5,56,23,0,0],"RequestId":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"Selector":null,"OppositeBridge":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"ChainId":94,"Raw":{"address":"0x0000000000000000000000000000000000000000","topics":null,"data":"0x","blockNumber":"0x0","transactionHash":"0x0000000000000000000000000000000000000000000000000000000000000000","transactionIndex":"0x0","blockHash":"0x0000000000000000000000000000000000000000000000000000000000000000","logIndex":"0x0","removed":false}}`
	jb, err := bridgeEvent2.ToJson()
	assert.NoError(t, err)
	assert.Equal(t, jbExpected, string(jb))

	// test DestChainId
	uChainId, fromHead := bridgeEvent2.DstChainId()
	assert.Equal(t, false, fromHead)
	assert.Equal(t, uint64(94), uChainId)
}
