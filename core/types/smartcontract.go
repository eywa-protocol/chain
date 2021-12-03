package types

import "github.com/eywa-protocol/chain/common"

type SmartCodeEvent struct {
	TxHash common.Uint256
	Action string
	Result interface{}
	Error  int64
}
