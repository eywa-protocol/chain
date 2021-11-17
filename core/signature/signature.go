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

	if !subSig.VerifyMultisig(allPub, subPub, data, big.NewInt(mask)) {
		return errors.New("Multisignature verification failed")
	}
	return nil
}
