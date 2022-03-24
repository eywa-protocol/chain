package payload

import (
	"math/rand"
	"reflect"
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
	reqId       [32]byte
	receiveSide ethCommon.Address
	bridgeFrom  [32]byte
	txHash      ethCommon.Hash
	x           ReceiveRequestEvent
)

func init() {
	rand.Read(reqId[:])
	rand.Read(receiveSide[:])
	rand.Read(bridgeFrom[:])
	rand.Read(txHash[:])

	x = ReceiveRequestEvent{
		OriginData: wrappers.BridgeReceiveRequest{
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
	}

}

func TestReceiveRequest_Borsh(t *testing.T) {
	data, err := borsh.Serialize(x)
	require.NoError(t, err)
	t.Log(data)
	y := new(ReceiveRequestEvent)
	err = borsh.Deserialize(y, data)
	require.True(t, reflect.DeepEqual(x, *y))
	sink := common.NewZeroCopySink(nil)
	err = x.Serialization(sink)
	assert.NoError(t, err)
	t.Log(data)
	t.Log(sink)
	// require.True(t, reflect.DeepEqual(data, sink))
	var bridgeEvent2 ReceiveRequestEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, x, bridgeEvent2)

	// test ToJson
	jbExpected := `{"ReqId":[3,124,77,123,187,4,7,209,226,198,73,129,133,90,216,104,29,13,134,209,233,30,0,22,121,57,203,102,148,210,196,34],"ReceiveSide":"0xacd208a0072939487f6999eb9d18a44784045d87","BridgeFrom":[243,198,124,242,39,70,233,149,175,90,37,54,121,81,186,162,255,108,212,113,196,131,241,95,185,11,173,179,124,88,33,182],"Raw":{"address":"0xacd208a0072939487f6999eb9d18a44784045d87","topics":null,"data":"0x","blockNumber":"0x0","transactionHash":"0xd95526a41a9504680b4e7c8b763a1b1d49d4955c8486216325253fec738dd7a9","transactionIndex":"0x0","blockHash":"0xd95526a41a9504680b4e7c8b763a1b1d49d4955c8486216325253fec738dd7a9","logIndex":"0x0","removed":false}}`
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
	assert.Equal(t, x, y)
}

func TestBridgeEvent_Serialization_ReceiveRequestEvent(t *testing.T) {
	sink := common.NewZeroCopySink(nil)
	err := x.Serialization(sink)
	assert.NoError(t, err)
	var recReqEvent2 ReceiveRequestEvent
	err = recReqEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, x, recReqEvent2)
}
