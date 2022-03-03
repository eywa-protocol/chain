package account

import (
	"github.com/eywa-protocol/bls-crypto/bls"
)

type Account struct {
	PrivateKey bls.PrivateKey
	PublicKey  bls.PublicKey
	Id         byte
}

func NewAccount(id byte) *Account {
	pri, pub := bls.GenerateRandomKey()
	return &Account{
		PrivateKey: pri,
		PublicKey:  pub,
		Id:         id,
	}
}

func (this *Account) PrivKey() bls.PrivateKey {
	return this.PrivateKey
}

func (this *Account) PubKey() bls.PublicKey {
	return this.PublicKey
}
