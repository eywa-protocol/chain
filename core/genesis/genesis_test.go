package genesis

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.InitLog(0, log.Stdout)
	m.Run()
	os.RemoveAll("./ActorLog")
}

func TestGenesisBlockInit(t *testing.T) {
	_, pub := bls.GenerateRandomKey()
	conf := &config.GenesisConfig{}
	block, err := BuildGenesisBlock([]bls.PublicKey{pub}, conf)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.NotEqual(t, block.Header.TransactionsRoot, common.UINT256_EMPTY)
}
