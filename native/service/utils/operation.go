package utils

import (
	"crypto"
	"encoding/binary"
	"fmt"
	"github.com/eywa-protocol/chain/common"
	"github.com/eywa-protocol/chain/native"
	"github.com/ontio/ontology-crypto/vrf"
)

func ConcatKey(contract common.Address, args ...[]byte) []byte {
	temp := contract[:]
	for _, arg := range args {
		temp = append(temp, arg...)
	}
	return temp
}

func ValidateOwner(native *native.NativeService, address common.Address) error {
	if native.CheckWitness(address) == false {
		return fmt.Errorf("validateOwner, authentication failed!")
	}
	return nil
}

func GetUint32Bytes(num uint32) []byte {
	var p [4]byte
	binary.LittleEndian.PutUint32(p[:], num)
	return p[:]
}

func GetBytesUint32(b []byte) uint32 {
	if len(b) != 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(b[:])
}

func GetBytesUint64(b []byte) uint64 {
	if len(b) != 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(b[:])
}

func GetUint64Bytes(num uint64) []byte {
	var p [8]byte
	binary.LittleEndian.PutUint64(p[:], num)
	return p[:]
}

func ValidatePeerPubKeyFormat(pubkey string) error {
	pk := crypto.PublicKey([]byte("pk"))
	if !vrf.ValidatePublicKey(pk) {
		return fmt.Errorf("invalid for VRF")
	}
	return nil
}
