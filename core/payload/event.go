package payload

import (
	"github.com/eywa-protocol/chain/common"
)

type RelayerEvent struct {
	OriginData CrossChainRequest
}

func (r RelayerEvent) Deserialization(source *common.ZeroCopySource) error {
	panic("implement me")
}

func (r RelayerEvent) Serialization(sink *common.ZeroCopySink) {
	panic("implement me")
}

type CrossChainRequest struct {
	RequestType    string
	BridgePubKey   []byte
	RequestId      []byte
	Selector       []byte
	ReceiveSide    []byte
	OppositeBridge []byte
	ChainId        uint64
	LogResult      struct{}
}
