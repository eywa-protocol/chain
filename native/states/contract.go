package states

import (
	"errors"
	"fmt"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/native/event"
)

const MAX_NATIVE_VERSION = 0

// Invoke smart contract struct
// Param Version: invoke smart contract version, default 0
// Param Address: invoke on blockchain smart contract by address
// Param Method: invoke smart contract method, default ""
// Param Args: invoke smart contract arguments
type ContractInvokeParam struct {
	Version byte
	Address common.Address
	Method  string
	Args    []byte
}

func (this *ContractInvokeParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteByte(this.Version)
	sink.WriteAddress(this.Address)
	sink.WriteVarBytes([]byte(this.Method))
	sink.WriteVarBytes([]byte(this.Args))
}

// `ContractInvokeParam.Args` has reference of `source`
func (this *ContractInvokeParam) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.Version, eof = source.NextByte()
	if eof {
		return errors.New("[ContractInvokeParam] deserialize version error")
	}
	if this.Version > MAX_NATIVE_VERSION {
		return fmt.Errorf("[ContractInvokeParam] current version %d over max native contract version %d", this.Version, MAX_NATIVE_VERSION)
	}

	this.Address, eof = source.NextAddress()
	if eof {
		return errors.New("[ContractInvokeParam] deserialize address error")
	}
	var method []byte
	method, eof = source.NextVarBytes()
	if eof {
		return errors.New("[ContractInvokeParam] deserialize method error")
	}
	this.Method = string(method)

	this.Args, eof = source.NextVarBytes()
	if eof {
		return errors.New("[ContractInvokeParam] deserialize args error")
	}
	return nil
}

type PreExecResult struct {
	State  byte
	Result interface{}
	Notify []*event.NotifyEventInfo
}
