package payload

import (
	"github.com/eywa-protocol/chain/common"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
	"testing"
)

func TestBridgeSolToEvmEvent_Serialize(t *testing.T) {
	bEvt := SolanaToEVMEvent{
		OriginData: bridge.BridgeEvent{struct {
			RequestType    string
			BridgePubKey   solana.PublicKey
			RequestId      solana.PublicKey
			Selector       []uint8
			ReceiveSide    [20]uint8
			OppositeBridge [20]uint8
			ChainId        uint64
			LogResult      ws.LogResult
		}{
			RequestType:    "test",
			BridgePubKey:   solana.PublicKey{},
			RequestId:      solana.PublicKey{},
			Selector:       []byte("testselector"),
			ReceiveSide:    common.Address{},
			OppositeBridge: common.Address{},
			ChainId:        uint64(3),
			LogResult: ws.LogResult{
				Context: struct {
					Slot uint64
				}{},
				Value: struct {
					Signature solana.Signature `json:"signature"`
					Err       interface{}      `json:"err"`
					Logs      []string         `json:"logs"`
				}{},
			}}},
	}

	sink := common.NewZeroCopySink(nil)
	bEvt.Serialization(sink)
	var bridgeEvent2 SolanaToEVMEvent
	err := bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bEvt, bridgeEvent2)
}
