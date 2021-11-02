/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/

package account

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
)

type Account struct {
	PrivateKey bls.PrivateKey
	PublicKey  bls.PublicKey
	Id         byte
	Address    common.Address
}

func NewAccount(id byte) *Account {
	pri, pub := bls.GenerateRandomKey()
	address := types.AddressFromPubKey(pub)
	return &Account{
		PrivateKey: pri,
		PublicKey:  pub,
		Id:         id,
		Address:    address,
	}
}

func (this *Account) PrivKey() bls.PrivateKey {
	return this.PrivateKey
}

func (this *Account) PubKey() bls.PublicKey {
	return this.PublicKey
}


