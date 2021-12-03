module github.com/eywa-protocol/chain

go 1.1

replace github.com/eywa-protocol/chain => ./

require (
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/ethereum/go-ethereum v1.10.12
	github.com/eywa-protocol/bls-crypto v0.1.2
	github.com/eywa-protocol/wrappers v0.0.4
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/itchyny/base58-go v0.1.0
	github.com/ontio/ontology-crypto v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
)
