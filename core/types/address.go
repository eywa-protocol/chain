package types

import (
	"fmt"
	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
)

func AddressFromPubKey(pubkey bls.PublicKey) common.Address {
	buf := pubkey.Marshal()
	return common.AddressFromBytes(buf)
}

func AddressFromMultiPubKeys(pubkeys []bls.PublicKey, m int) (common.Address, error) {
	sink := common.NewZeroCopySink(nil)
	if err := EncodeMultiPubKeyProgramInto(sink, pubkeys, uint16(m)); err != nil {
		return common.ADDRESS_EMPTY, err
	}
	fmt.Printf("\nsink.Bytes() %v", common.ToHexString(sink.Bytes()))
	add := common.AddressFromBytes(sink.Bytes())
	fmt.Printf("\nadd %v", add)
	return add, nil
}

func AddressFromPubLeySlice(bookkeepers []bls.PublicKey) (common.Address, error) {
	if len(bookkeepers) == 1 {
		return AddressFromPubKey(bookkeepers[0]), nil
	}
	return AddressFromMultiPubKeys(bookkeepers, len(bookkeepers)-(len(bookkeepers)-1)/3)
}
