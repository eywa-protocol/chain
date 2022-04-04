module github.com/eywa-protocol/chain

go 1.16

require (
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/ethereum/go-ethereum v1.10.13
	github.com/eywa-protocol/bls-crypto v0.1.3
	github.com/eywa-protocol/wrappers v0.2.16
	github.com/gagliardetto/solana-go v1.0.2
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/itchyny/base58-go v0.1.0
	github.com/near/borsh-go v0.3.1-0.20210831082424-4377deff6791
	github.com/ontio/ontology-crypto v1.2.1
	github.com/sirupsen/logrus v1.2.0
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	gitlab.digiu.ai/blockchainlaboratory/eywa-solana v1.2.2
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
)

// replace gitlab.digiu.ai/blockchainlaboratory/eywa-solana => ../solana/
// replace github.com/eywa-protocol/wrappers => ../wrappers
