package payload

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
)

func TestBridgeSolanaEvent_Serialize(t *testing.T) {
	bridgeEvent := BridgeSolanaEvent{
		OriginData: wrappers.BridgeOracleRequestSolana{
			RequestType: "setRequest",
			Bridge:      [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 90, 1, 2, 3, 4, 5, 6, 7, 78, 9, 0, 1, 2, 2, 3, 43, 4, 4, 5, 5, 56, 23},
			Chainid:     big.NewInt(94),
		}}

	sink := common.NewZeroCopySink(nil)
	bridgeEvent.Serialization(sink)
	var bridgeEvent2 BridgeSolanaEvent
	err := bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bridgeEvent, bridgeEvent2)
}
