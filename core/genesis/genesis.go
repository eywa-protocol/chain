package genesis

import (
	"fmt"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/eywa-protocol/chain/native/service/utils"
	"github.com/eywa-protocol/chain/native/states"
)

const (
	INIT_CONFIG = "initConfig"
)

var GenesisEpochValidators []bls.PublicKey

// BuildGenesisBlock returns the genesis block with default consensus bookkeeper list
func BuildGenesisBlock(defaultEpoch []bls.PublicKey) (*types.Block, error) {

	GenesisEpochValidators = defaultEpoch
	nextEpoch, err := types.AddressFromPubLeySlice(defaultEpoch)
	if err != nil {
		return nil, fmt.Errorf("[Block],BuildGenesisBlock err with GetEpochAddress: %s", err)
	}
	nodeManagerConfig := newNodeManagerEpochInit([]byte(nextEpoch.ToHexString()))

	genesisHeader := &types.Header{
		ChainID:          0,
		PrevBlockHash:    common.Uint256{},
		EpochBlockHash:   common.Uint256{},
		TransactionsRoot: common.Uint256{},
		SourceHeight:     0,
		Height:           0,
	}

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			nodeManagerConfig,
		},
	}
	genesisBlock.RebuildMerkleRoot()
	return genesisBlock, nil
}

func newNodeManagerInit(config []byte) *types.Transaction {
	tx, err := NewInitNodeManagerTransaction(config)
	if err != nil {
		panic("construct genesis node manager transaction error ")
	}
	return tx
}

func newNodeManagerEpochInit(config []byte) *types.Transaction {
	tx, err := NewInitNodeManagerEpochTransaction(config)
	if err != nil {
		panic("construct genesis node manager transaction error ")
	}
	if (&types.Transaction{} == tx) {
		panic("empty transaction")
	}
	if tx.Payload == nil {
		panic("transaction payload is NIL !")
	}
	return tx
}

//NewInvokeTransaction return smart contract invoke transaction
func NewInvokeTransaction(invokeCode []byte, nonce uint32) *types.Transaction {
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	tx := &types.Transaction{
		TxType:  types.Invoke,
		Payload: invokePayload,
	}

	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	if err != nil {
		return &types.Transaction{}
	}
	tx, err = types.TransactionFromRawBytes(sink.Bytes())
	return tx
}

//NewInvokeTransaction return smart contract invoke transaction
func NewEpochTransaction(invokeCode []byte, nonce uint32) *types.Transaction {

	tx := &types.Transaction{
		TxType:  types.Epoch,
		Payload: &payload.Epoch{Data: invokeCode},
	}

	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	if err != nil {
		return &types.Transaction{}
	}
	tx, err = types.TransactionFromRawBytes(sink.Bytes())
	return tx
}

func NewInitNodeManagerTransaction(
	paramBytes []byte,
) (*types.Transaction, error) {
	contractInvokeParam := &states.ContractInvokeParam{Address: utils.NodeManagerContractAddress,
		Method: INIT_CONFIG, Args: paramBytes}
	invokeCode := new(common.ZeroCopySink)
	contractInvokeParam.Serialization(invokeCode)
	return NewInvokeTransaction(invokeCode.Bytes(), 0), nil
}

func NewInitNodeManagerEpochTransaction(
	paramBytes []byte,
) (*types.Transaction, error) {

	return NewEpochTransaction(paramBytes, 0), nil
}
