package genesis

import (
	"fmt"
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/common/constants"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/chain/core/types"
	"github.com/eywa-protocol/chain/native/service/utils"
	"github.com/eywa-protocol/chain/native/states"
)

const (
	GenesisNonce uint64 = 1

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
	consensusPayload := []byte("0")

	genesisHeader := &types.Header{
		ChainID:          0,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        constants.GENESIS_BLOCK_TIMESTAMP,
		Height:           uint32(0),
		ConsensusData:    GenesisNonce,
		NextEpoch:        nextEpoch,
		ConsensusPayload: consensusPayload,
		BlockRoot:        common.UINT256_EMPTY,
		EpochValidators:  GenesisEpochValidators,
		SigData:          nil,
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
		Nonce:   nonce,
		ChainID: 0,
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
		Nonce:   nonce,
		ChainID: 0,
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
