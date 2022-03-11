package native

import (
	"fmt"

	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/eywa-protocol/chain/merkle"
	"github.com/eywa-protocol/chain/native/event"
	"github.com/eywa-protocol/chain/native/states"
	"github.com/eywa-protocol/chain/native/storage"
)

type (
	Handler         func(native *NativeService) ([]byte, error)
	RegisterService func(native *NativeService)
)

var (
	Contracts = make(map[common.Address]RegisterService)
)

const (
	MAX_CONTEXT_LEN = 1024
)

// Native service struct
// Invoke a native smart contract, new a native service
type NativeService struct {
	cacheDB       *storage.CacheDB
	chainID       uint64
	serviceMap    map[string]Handler
	notifications []*event.NotifyEventInfo
	input         []byte
	tx            types.Transaction
	height        uint64
	blockHash     common.Uint256
	crossHashes   []common.Uint256
	contexts      []common.Address
	preExec       bool
}

func NewNativeService(cacheDB *storage.CacheDB, tx types.Transaction,
	height uint64, blockHash common.Uint256, chainID uint64, input []byte, preExec bool) (*NativeService, error) {
	// if tx.ChainID != chainID {
	// 	return nil, fmt.Errorf("NewNativeService Error: ChainId in Tx not equal to current block ChainId, expect: %d, got: %d", chainID, tx.ChainID)
	// }
	service := &NativeService{
		cacheDB:    cacheDB,
		tx:         tx,
		height:     height,
		blockHash:  blockHash,
		serviceMap: make(map[string]Handler),
		input:      input,
		chainID:    chainID,
		preExec:    preExec,
	}

	return service, nil
}

func (this *NativeService) Register(methodName string, handler Handler) {
	this.serviceMap[methodName] = handler
}

func (this *NativeService) Invoke() (interface{}, error) {
	invokeParam := new(states.ContractInvokeParam)
	if err := invokeParam.Deserialization(common.NewZeroCopySource(this.input)); err != nil {
		return nil, err
	}
	services, ok := Contracts[invokeParam.Address]
	if !ok {
		return false, fmt.Errorf("[Invoke] Native contract address %x haven't been registered.", invokeParam.Address)
	}
	services(this)
	service, ok := this.serviceMap[invokeParam.Method]
	if !ok {
		return false, fmt.Errorf("[Invoke] Native contract %x doesn't support this function %s.",
			invokeParam.Address, invokeParam.Method)
	}
	args := this.input
	this.input = invokeParam.Args
	notifications := this.notifications
	this.notifications = []*event.NotifyEventInfo{}
	hashes := this.crossHashes
	this.crossHashes = []common.Uint256{}
	if err := this.PushContext(invokeParam.Address); err != nil {
		return err, nil
	}
	result, err := service(this)
	if err != nil {
		return result, fmt.Errorf("[Invoke] Native serivce function execute error:%s", err)
	}
	this.PopContext()
	this.notifications = append(notifications, this.notifications...)
	this.crossHashes = append(this.crossHashes, hashes...)
	this.input = args
	return result, nil
}

func (this *NativeService) NativeCall(address common.Address, method string, args []byte) (interface{}, error) {
	c := states.ContractInvokeParam{
		Address: address,
		Method:  method,
		Args:    args,
	}
	sink := common.NewZeroCopySink(nil)
	c.Serialization(sink)
	this.input = sink.Bytes()
	return this.Invoke()
}

func (this *NativeService) PushContext(address common.Address) error {
	if len(this.contexts) > MAX_CONTEXT_LEN {
		return fmt.Errorf("context over max context lenght:%d max contexts lenght:%d", len(this.contexts), MAX_CONTEXT_LEN)
	}
	this.contexts = append(this.contexts, address)
	return nil
}

func (this *NativeService) CallingContext() common.Address {
	if len(this.contexts) < 2 {
		return common.ADDRESS_EMPTY
	}
	return this.contexts[len(this.contexts)-2]
}

func (this *NativeService) PopContext() {
	if len(this.contexts) > 1 {
		this.contexts = this.contexts[:len(this.contexts)-1]
	}
}

func (this *NativeService) CurrentContext() common.Address {
	if len(this.contexts) == 0 {
		return common.ADDRESS_EMPTY
	}
	return this.contexts[len(this.contexts)-1]
}

func (this *NativeService) PutMerkleVal(data []byte) {
	this.crossHashes = append(this.crossHashes, merkle.HashLeaf(data))
}

/*func (this *NativeService) checkAccountAddress(address common.Address) bool {
	addresses, err := this.tx.GetSignatureAddresses()
	if err != nil {
		log.Errorf("get signature address error:%v", err)
		return false
	}
	for _, v := range addresses {
		if v == address {
			return true
		}
	}
	return false
}*/

func (this *NativeService) checkContractAddress(address common.Address) bool {
	if this.CallingContext() != common.ADDRESS_EMPTY && this.CallingContext() == address {
		return true
	}
	return false
}

// CheckWitness check whether authorization correct
func (this *NativeService) CheckWitness(address common.Address) bool {
	if /*this.checkAccountAddress(address) ||*/ this.checkContractAddress(address) {
		return true
	}
	return false
}

func (this *NativeService) AddNotify(notify *event.NotifyEventInfo) {
	this.notifications = append(this.notifications, notify)
}

func (this *NativeService) GetCacheDB() *storage.CacheDB {
	return this.cacheDB
}

func (this *NativeService) GetInput() []byte {
	return this.input
}

func (this *NativeService) GetTx() types.Transaction {
	return this.tx
}

func (this *NativeService) GetHeight() uint64 {
	return this.height
}

func (this *NativeService) GetChainID() uint64 {
	return this.chainID
}

func (this *NativeService) GetNotify() []*event.NotifyEventInfo {
	return this.notifications
}

func (this *NativeService) GetCrossHashes() []common.Uint256 {
	return this.crossHashes
}
