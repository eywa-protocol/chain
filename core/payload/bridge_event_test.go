package payload

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestBridgeEvent_Serialize(t *testing.T) {
	bridgeEvent := BridgeEvent{
		OriginData: wrappers.BridgeOracleRequest{
			RequestType: "setRequest",
			Bridge:      ethCommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
			Chainid:     big.NewInt(94),
		}}

	sink := common.NewZeroCopySink(nil)
	bridgeEvent.Serialization(sink)
	var bridgeEvent2 BridgeEvent
	err := bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bridgeEvent, bridgeEvent2)
}
