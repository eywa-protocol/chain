/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/

package types

import "gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"

type SmartCodeEvent struct {
	TxHash common.Uint256
	Action string
	Result interface{}
	Error  int64
}
