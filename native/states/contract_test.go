/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/
package states

import (
	"testing"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
)

func TestContract_Serialize_Deserialize(t *testing.T) {
	addr := common.AddressFromBytes([]byte{1})

	c := &ContractInvokeParam{
		Version: 0,
		Address: addr,
		Method:  "init",
		Args:    []byte{2},
	}
	sink := common.NewZeroCopySink(nil)
	c.Serialization(sink)

	v := new(ContractInvokeParam)
	if err := v.Deserialization(common.NewZeroCopySource(sink.Bytes())); err != nil {
		t.Fatalf("ContractInvokeParam deserialize error: %v", err)
	}
}
