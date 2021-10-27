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

package signature

import (
	"errors"
	"math/big"

	"github.com/eywa-protocol/bls-crypto/bls"
)

// Sign returns the signature of data using privKey
func Sign(signer Signer, data []byte) ([]byte, error) {
	sig, _ := signAndGetSig(signer, data)
	return sig.Marshal(), nil
}

func signAndGetSig(signer Signer, data []byte) (bls.Signature, error) {
	prKey := signer.PrivKey()
	return prKey.Sign(data), nil
}

func Signature(signer Signer, data []byte) (bls.Signature, error) {
	return signAndGetSig(signer, data)
}

// Verify check the signature of data using pubKey
func Verify(pubKey bls.PublicKey, data []byte, signature bls.Signature) error {

	// sigObj, err := bls.UnmarshalSignature(signature)

	// if err != nil {
	// 	return errors.New("invalid signature data: " + err.Error())
	// }

	if !signature.Verify(pubKey, data) {
		return errors.New("signature verification failed")
	}

	return nil
}

// VerifyMultiSignature check whether more than m sigs are signed by the keys
func VerifyMultiSignature(data []byte, subSig bls.Signature, allPub bls.PublicKey, subPub bls.PublicKey, mask int64) error {
	// TODO resore VerifyMultiSignature with bls
	//n := len(keys)
	//
	//if len(sigs) < m {
	//	return errors.New("not enough signatures in multi-signature")
	//}
	//
	//mask := make([]bool, n)
	//for i := 0; i < m; i++ {
	//	valid := false
	//
	//	//sig, err := s.Deserialize(sigs[i])
	//	if err != nil {
	//		return errors.New("invalid signature data")
	//	}
	//	for j := 0; j < n; j++ {
	//		if mask[j] {
	//			continue
	//		}
	//		if Verify(keys[j], data, sig) {
	//			mask[j] = true
	//			valid = true
	//			break
	//		}
	//	}
	//
	//	if valid == false {
	//		return errors.New("multi-signature verification failed")
	//	}
	//}

	if !subSig.VerifyMultisig(allPub, subPub, data, big.NewInt(mask)) {
		return errors.New("Multisignature verification failed")
	}
	return nil
}
