package signature

import (
	"github.com/eywa-protocol/bls-crypto/bls"
)

// Signer is the abstract interface of user's information(Keys) for signing data.
type Signer interface {
	//get signer's private key
	PrivKey() bls.PrivateKey

	//get signer's public key
	PubKey() bls.PublicKey

	//Scheme() signature.SignatureScheme
}
