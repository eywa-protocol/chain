package types

import (
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/core/payload"
	"github.com/eywa-protocol/wrappers"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-solana/sdk/bridge"
)

func Test_HeaderMarshal(t *testing.T) {
	hash := common.Uint256{0xCA, 0xFE, 0xBA, 0xBE}

	header := Header{
		ChainID:          1111,
		PrevBlockHash:    hash,
		EpochBlockHash:   hash,
		TransactionsRoot: hash,
		SourceHeight:     100,
		Height:           10,
		Signature:        bls.NewZeroMultisig(),
	}

	sink := common.NewZeroCopySink(nil)
	err := header.Serialization(sink)
	assert.NoError(t, err)
	// t.Log(sink.Bytes())

	var received Header
	source := common.NewZeroCopySource(sink.Bytes())
	err = received.Deserialization(source)
	assert.NoError(t, err)
	assert.Equal(t, header, received)
}

func Test_EmptyBlockMarshal(t *testing.T) {
	hash := common.Uint256{0xCA, 0xFE, 0xBA, 0xBE}
	block := NewBlock(1111, hash, hash, 100, 10, Transactions{})

	sink := common.NewZeroCopySink(nil)
	err := block.Serialization(sink)
	assert.NoError(t, err)
	// t.Log(sink.Bytes())

	var received Block
	source := common.NewZeroCopySource(sink.Bytes())
	err = received.Deserialization(source)
	assert.NoError(t, err)
	assert.Equal(t, *block, received)
}

func Test_BlockMarshal(t *testing.T) {
	hash := common.Uint256{0xCA, 0xFE, 0xBA, 0xBE}

	txs := make(Transactions, 0)
	{
		tx := payload.NewBridgeEvent(&wrappers.BridgeOracleRequest{
			RequestType: "setRequest",
			Bridge:      ethcommon.HexToAddress("0x0c760E9A85d2E957Dd1E189516b6658CfEcD3985"),
			ChainId:     big.NewInt(1111),
		})
		txs = append(txs, ToTransaction(tx))
	}
	{
		tx := &payload.ReceiveRequestEvent{
			OriginData: wrappers.BridgeReceiveRequest{
				ReqId:       payload.RequestId{1, 2, 3, 4, 5},
				ReceiveSide: ethcommon.Address{6, 7, 8, 9, 10},
				BridgeFrom:  [32]byte{11, 12, 13, 14, 15},
			},
		}
		txs = append(txs, ToTransaction(tx))
	}
	{
		tx := payload.NewBridgeSolanaEvent(&wrappers.BridgeOracleRequestSolana{
			RequestType: "setRequest",
			Bridge:      [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 90, 1, 2, 3, 4, 5, 6, 7, 78, 9, 0, 1, 2, 2, 3, 43, 4, 4, 5, 5, 56, 23},
			ChainId:     big.NewInt(1111),
		})
		txs = append(txs, ToTransaction(tx))
	}
	{
		tx := &payload.SolanaToEVMEvent{
			OriginData: bridge.BridgeEvent{
				OracleRequest: bridge.OracleRequest{
					RequestType:    "test",
					BridgePubKey:   solana.PublicKey{},
					RequestId:      solana.PublicKey{1, 2, 3, 4, 5},
					Selector:       []byte("testselector"),
					ReceiveSide:    common.Address{},
					OppositeBridge: common.Address{},
					ChainId:        uint64(3),
				},
				Signature: solana.Signature{},
				Slot:      uint64(3),
			},
		}
		txs = append(txs, ToTransaction(tx))
	}

	block := NewBlock(1111, hash, hash, 100, 10, txs)
	block.Hash()
	t.Logf("Transactions: %d, Block hash: %x", len(block.Transactions), block.Hash())

	sink := common.NewZeroCopySink(nil)
	err := block.Serialization(sink)
	assert.NoError(t, err)
	// t.Log(sink.Bytes())

	var received Block
	source := common.NewZeroCopySource(sink.Bytes())
	err = received.Deserialization(source)
	assert.NoError(t, err)

	// Compare blocks
	assert.Equal(t, block.Header, received.Header)
	assert.Equal(t, len(block.Transactions), len(received.Transactions))
	for i := range block.Transactions {
		assert.Equal(t, block.Transactions[i].Payload.Data(), received.Transactions[i].Payload.Data())
	}

	// text, _ := json.Marshal(block)
	// t.Logf(string(text))
}
