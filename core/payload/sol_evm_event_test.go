package payload

import (
	"testing"

	"github.com/eywa-protocol/chain/common"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
)

func TestBridgeSolToEvmEvent_Serialize(t *testing.T) {
	bEvt := SolanaToEVMEvent{
		OriginData: bridge.BridgeEvent{
			OracleRequest: bridge.OracleRequest{
				RequestType:    "test",
				BridgePubKey:   solana.PublicKey{},
				RequestId:      solana.PublicKey{},
				Selector:       []byte("testselector"),
				ReceiveSide:    common.Address{},
				OppositeBridge: common.Address{},
				ChainId:        uint64(3),
			},
			Signature: solana.Signature{},
			Slot:      uint64(3),
		},
	}

	sink := common.NewZeroCopySink(nil)
	bEvt.Serialization(sink)
	// t.Log(sink.Bytes())

	var bridgeEvent2 SolanaToEVMEvent
	err := bridgeEvent2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, bEvt, bridgeEvent2)
}
