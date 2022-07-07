package payload

import (
	"math/rand"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
	"github.com/near/borsh-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	reqId       RequestId
	receiveSide ethCommon.Address
	bridgeFrom  [32]byte
	txHash      ethCommon.Hash
	x           *ReceiveRequestEvent
)

func init() {
	rand.Read(reqId[:])
	rand.Read(receiveSide[:])
	rand.Read(bridgeFrom[:])
	rand.Read(txHash[:])

	x = NewReceiveRequestEvent(&wrappers.BridgeReceiveRequest{
		ReqId:       reqId,
		ReceiveSide: receiveSide,
		BridgeFrom:  bridgeFrom,
		Raw: types.Log{
			Address:     receiveSide,
			Topics:      nil,
			Data:        nil,
			BlockNumber: 0,
			TxHash:      txHash,
			TxIndex:     0,
			BlockHash:   txHash,
			Index:       0,
			Removed:     false,
		},
	},
	)

}

func TestReceiveRequest_Borsh(t *testing.T) {
	data, err := borsh.Serialize(x.Data())
	require.NoError(t, err)
	t.Log(data)
	y := new(ReceiveRequestEventData)
	err = borsh.Deserialize(y, data)
	require.Equal(t, x.Data(), *y)
	sink := common.NewZeroCopySink(nil)
	err = x.Serialization(sink)
	assert.NoError(t, err)
	t.Log(data)
	t.Log(sink)
	// require.True(t, reflect.DeepEqual(data, sink))
	var bridgeEvent2 ReceiveRequestEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, x.Data(), bridgeEvent2.Data())

	// test ToJson
	jbExpected := `{"ReqId":"037c4d7bbb0407d1e2c64981855ad8681d0d86d1e91e00167939cb6694d2c422","ReceiveSide":"0xacd208a0072939487f6999eb9d18a44784045d87","BridgeFrom":"f3c67cf22746e995af5a25367951baa2ff6cd471c483f15fb90badb37c5821b6"}`
	jb, err := bridgeEvent2.ToJson()
	assert.NoError(t, err)
	assert.Equal(t, jbExpected, string(jb))

	// test DestChainId
	uChainId, fromHead := bridgeEvent2.DstChainId()
	assert.Equal(t, true, fromHead)
	assert.Equal(t, uint64(0), uChainId)
}

func TestBridgeEvent_Serialize2(t *testing.T) {
	sink := common.NewZeroCopySink(nil)
	err := x.Serialization(sink)
	assert.NoError(t, err)
	var y ReceiveRequestEvent
	err = y.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, x.Data(), y.Data())
}

func TestBridgeEvent_Serialization_ReceiveRequestEvent(t *testing.T) {
	sink := common.NewZeroCopySink(nil)
	err := x.Serialization(sink)
	assert.NoError(t, err)
	var recReqEvent2 ReceiveRequestEvent
	err = recReqEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, x.Data(), recReqEvent2.Data())
}
