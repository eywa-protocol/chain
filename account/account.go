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

package account

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/core/types"
	//"github.com/ethereum/go-ethereum/common"
)

/* crypto object */
type Account struct {
	PrivateKey bls.PrivateKey
	PublicKey  bls.PublicKey
	Id         byte
	Address    common.Address
	//SigScheme  s.SignatureScheme
	//URL        accounts.URL
}

func NewAccount(id byte) *Account {
	// Determine the public key algorithm and parameters according to
	// the encrypt.
	//var pkAlgorithm keypair.KeyType
	//var params interface{}
	//var scheme s.SignatureScheme
	//var err error
	//if "" != encrypt {
	//	scheme, err = s.GetScheme(encrypt)
	//} else {
	//	scheme = s.SHA256withECDSA
	//}
	//if err != nil {
	//	log.Warn("unknown signature scheme, use SHA256withECDSA as default.")
	//	scheme = s.SHA256withECDSA
	//}
	//switch scheme {
	//case s.SHA224withECDSA, s.SHA3_224withECDSA:
	//	pkAlgorithm = keypair.PK_ECDSA
	//	params = keypair.P224
	//case s.SHA256withECDSA, s.SHA3_256withECDSA, s.RIPEMD160withECDSA:
	//	pkAlgorithm = keypair.PK_ECDSA
	//	params = keypair.P256
	//case s.SHA384withECDSA, s.SHA3_384withECDSA:
	//	pkAlgorithm = keypair.PK_ECDSA
	//	params = keypair.P384
	//case s.SHA512withECDSA, s.SHA3_512withECDSA:
	//	pkAlgorithm = keypair.PK_ECDSA
	//	params = keypair.P521
	//case s.SM3withSM2:
	//	pkAlgorithm = keypair.PK_SM2
	//	params = keypair.SM2P256V1
	//case s.SHA512withEDDSA:
	//	pkAlgorithm = keypair.PK_EDDSA
	//	params = keypair.ED25519
	//}

	pri, pub := bls.GenerateRandomKey()
	address := types.AddressFromPubKey(pub)
	return &Account{
		PrivateKey: pri,
		PublicKey:  pub,
		Id:         id,
		Address:    address,
		//SigScheme:  scheme,
	}
}

func (this *Account) PrivKey() bls.PrivateKey {
	return this.PrivateKey
}

func (this *Account) PubKey() bls.PublicKey {
	return this.PublicKey
}

/*func (this *Account) Scheme() s.SignatureScheme {
	return this.SigScheme
}*/

//AccountMetadata all account info without private key
type AccountMetadata struct {
	IsDefault bool   //Is default account
	Label     string //Lable of account
	KeyType   string //KeyType ECDSA,SM2 or EDDSA
	Curve     string //Curve of key type
	Address   string //Address(base58) of account
	PubKey    string //Public  key
	SigSch    string //Signature scheme
	Salt      []byte //Salt
	Key       []byte //PrivateKey in encrypted
	EncAlg    string //Encrypt alg of private key
	Hash      string //Hash alg
}
