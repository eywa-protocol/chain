/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */
package genesis

import (
	"github.com/ontio/ontology-crypto/keypair"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-blockchain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-blockchain/common/config"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-blockchain/common/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.InitLog(0, log.Stdout)
	m.Run()
	os.RemoveAll("./ActorLog")
}

func TestGenesisBlockInit(t *testing.T) {
	_, pub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	conf := &config.GenesisConfig{}
	block, err := BuildGenesisBlock([]keypair.PublicKey{pub}, conf)
	assert.Nil(t, err)
	assert.NotNil(t, block)
	assert.NotEqual(t, block.Header.TransactionsRoot, common.UINT256_EMPTY)
}
