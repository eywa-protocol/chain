package genesis

import (
	"fmt"
	"github.com/eywa-protocol/bls-crypto/bls"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/native/service/utils"
	"time"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/constants"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/payload"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/native/states"
)

const (
	BlockVersion uint32 = 0
	GenesisNonce uint64 = 2083236893

	INIT_CONFIG = "initConfig"
)

var GenBlockTime = (config.DEFAULT_GEN_BLOCK_TIME * time.Second)

var GenesisBookkeepers []bls.PublicKey

// BuildGenesisBlock returns the genesis block with default consensus bookkeeper list
func BuildGenesisBlock(defaultBookkeeper []bls.PublicKey) (*types.Block, error) {

	GenesisBookkeepers = defaultBookkeeper
	nextBookkeeper, err := types.AddressFromPubLeySlice(defaultBookkeeper)
	if err != nil {
		return nil, fmt.Errorf("[Block],BuildGenesisBlock err with GetBookkeeperAddress: %s", err)
	}
	nodeManagerConfig := newNodeManagerEpochInit([]byte(nextBookkeeper.ToHexString()))
	consensusPayload := []byte("0")
	if err != nil {
		return nil, fmt.Errorf("consensus genesis init failed: %s", err)
	}

	genesisHeader := &types.Header{
		ChainID:          config.GetChainIdByNetId(config.DefConfig.P2PNode.NetworkId),
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        constants.GENESIS_BLOCK_TIMESTAMP,
		Height:           uint32(0),
		ConsensusData:    GenesisNonce,
		NextBookkeeper:   nextBookkeeper,
		ConsensusPayload: consensusPayload,
		BlockRoot:        common.UINT256_EMPTY,
		Bookkeepers:      nil,
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
		ChainID: config.GetChainIdByNetId(config.DefConfig.P2PNode.NetworkId),
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
		Payload: &payload.Epoch{Code: invokeCode},
		Nonce:   nonce,
		ChainID: config.GetChainIdByNetId(config.DefConfig.P2PNode.NetworkId),
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
	contractInvokeParam := &states.ContractInvokeParam{Address: utils.NodeManagerContractAddress,
		Method: INIT_CONFIG, Args: paramBytes}
	invokeCode := new(common.ZeroCopySink)
	contractInvokeParam.Serialization(invokeCode)

	return NewEpochTransaction(invokeCode.Bytes(), 0), nil
}
