package payload

import (
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
	"github.com/stretchr/testify/assert"
)

func TestBridgeEvent_Serialize(t *testing.T) {
	chainId := big.NewInt(94)
	bridgeEvent := NewBridgeEvent(&wrappers.BridgeOracleRequest{
		RequestType: "setRequest",
		Bridge:      ethCommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
		ChainId:     chainId,
	})

	sink := common.NewZeroCopySink(nil)
	err := bridgeEvent.Serialization(sink)
	assert.NoError(t, err)
	var bridgeEvent2 BridgeEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bridgeEvent.Data(), bridgeEvent2.Data())

	// test ToJson
	jbExpected := `{"RequestType":"setRequest","Bridge":"0x0c760e9a85d2e957dd1e189516b6658cfecd3985","ReqId":"0000000000000000000000000000000000000000000000000000000000000000","Selector":null,"ReceiveSide":"0x0000000000000000000000000000000000000000","OppositeBridge":"0x0000000000000000000000000000000000000000","ChainId":94}`
	jb, err := bridgeEvent2.ToJson()
	assert.NoError(t, err)
	assert.Equal(t, jbExpected, string(jb))

	// test DestChainId
	uChainId, fromHead := bridgeEvent.DstChainId()
	assert.Equal(t, false, fromHead)
	assert.Equal(t, chainId.Uint64(), uChainId)
}

func TestBridgeEvent_SerializeBorsh(t *testing.T) {
	bridgeEvent := NewBridgeEvent(&wrappers.BridgeOracleRequest{
		RequestType: "setRequest",
		Bridge:      ethCommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
		ChainId:     big.NewInt(94),
	})

	sink := common.NewZeroCopySink(nil)
	err := bridgeEvent.Serialization(sink)
	assert.NoError(t, err)
	var bridgeEvent2 BridgeEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bridgeEvent.Data(), bridgeEvent2.Data())
}
