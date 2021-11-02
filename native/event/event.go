package event

import (
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
)

const (
	EVENT_NOTIFY = "Notify"
)

// PushSmartCodeEvent push event content to socket.io
func PushSmartCodeEvent(txHash common.Uint256, errCode int64, action string, result interface{}) {
	//if events.DefActorPublisher == nil {
	//	return
	//}
	//smartCodeEvt := &types.SmartCodeEvent{
	//	TxHash: txHash,
	//	Action: action,
	//	Result: result,
	//	Error:  errCode,
	//}
	//events.DefActorPublisher.Publish(message.TOPIC_SMART_CODE_EVENT, &message.SmartCodeEventMsg{smartCodeEvt})
}
