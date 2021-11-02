package vote

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/genesis"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
)

func GetValidators(txs []*types.Transaction) ([]bls.PublicKey, error) {
	// TODO implement vote
	return genesis.GenesisBookkeepers, nil
}
