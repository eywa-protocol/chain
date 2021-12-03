package vote

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/core/genesis"
	"github.com/eywa-protocol/chain/core/types"
)

func GetValidators(txs []*types.Transaction) ([]bls.PublicKey, error) {
	// TODO implement vote
	return genesis.GenesisEpochValidators, nil
}
