package payload

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/wrappers"
	"github.com/near/borsh-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"reflect"
	"testing"
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
	x.Serialization(sink)
	t.Log(data)
	t.Log(sink)
	//require.True(t, reflect.DeepEqual(data, sink))
	var bridgeEvent2 ReceiveRequestEvent
	err = bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, x, bridgeEvent2)

}

func TestBridgeEvent_Serialize2(t *testing.T) {

	sink := common.NewZeroCopySink(nil)
	x.Serialization(sink)
	var y ReceiveRequestEvent
	err := y.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, x, y)
}

func TestBridgeEvent_Serialization_ReceiveRequestEvent(t *testing.T) {
	sink := common.NewZeroCopySink(nil)
	x.Serialization(sink)
	var recReqEvent2 ReceiveRequestEvent
	err := recReqEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, x, recReqEvent2)
}
