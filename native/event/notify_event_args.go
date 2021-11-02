/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/

package event

import (
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
)

const (
	CONTRACT_STATE_FAIL    byte = 0
	CONTRACT_STATE_SUCCESS byte = 1
)

// NotifyEventInfo describe smart contract event notify info struct
type NotifyEventInfo struct {
	ContractAddress common.Address
	States          interface{}
}

type ExecuteNotify struct {
	TxHash      common.Uint256
	State       byte
	GasConsumed uint64
	Notify      []*NotifyEventInfo
}
