package payload

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
)

func TestBridgeSolanaEvent_Serialize(t *testing.T) {
	bridgeEvent := NewBridgeSolanaEvent(&wrappers.BridgeOracleRequestSolana{
		RequestType: "setRequest",
		Bridge:      [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 90, 1, 2, 3, 4, 5, 6, 7, 78, 9, 0, 1, 2, 2, 3, 43, 4, 4, 5, 5, 56, 23},
		ChainId:     big.NewInt(94),
	})

	sink := common.NewZeroCopySink(nil)
	err := bridgeEvent.Serialization(sink)
	assert.NoError(t, err)
	var bridgeEvent2 BridgeSolanaEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bridgeEvent.Data(), bridgeEvent2.Data())

	// test ToJson
	jbExpected := `{"RequestType":"setRequest","Bridge":"01020304050607085a010203040506074e0900010202032b0404050538170000","ReqId":"0000000000000000000000000000000000000000000000000000000000000000","Selector":null,"OppositeBridge":"0000000000000000000000000000000000000000000000000000000000000000","ChainId":94}`
	jb, err := bridgeEvent2.ToJson()
	assert.NoError(t, err)
	assert.Equal(t, jbExpected, string(jb))

	// test DestChainId
	uChainId, fromHead := bridgeEvent2.DstChainId()
	assert.Equal(t, false, fromHead)
	assert.Equal(t, uint64(94), uChainId)
}
