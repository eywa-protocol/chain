module gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain

go 1.1

replace gitlab.digiu.ai/blockchainlaboratory/wrappers => ../eth-contracts/wrappers/

replace gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain => ./

require (
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/davecgh/go-spew v1.1.1
	github.com/ethereum/go-ethereum v1.10.8
	github.com/eywa-protocol/bls-crypto v0.1.2
	github.com/google/uuid v1.1.5
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/itchyny/base58-go v0.1.0
	github.com/ontio/ontology-crypto v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20210305035536-64b5b1c73954
	gitlab.digiu.ai/blockchainlaboratory/wrappers v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
)
