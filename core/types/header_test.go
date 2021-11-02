package types

import (
	"bytes"
	"fmt"
	"github.com/eywa-protocol/bls-crypto/bls"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
)

func TestHeader_Serialize(t *testing.T) {
	header := Header{}
	header.Height = 321
	header.Bookkeepers = make([]bls.PublicKey, 0)
	header.SigData = make([][]byte, 0)
	sink := common.NewZeroCopySink(nil)
	header.Serialization(sink)
	bs := sink.Bytes()

	var h2 Header
	source := common.NewZeroCopySource(bs)
	err := h2.Deserialization(source)
	assert.Equal(t, fmt.Sprint(header), fmt.Sprint(h2))

	assert.Nil(t, err)
}

func TestHeader(t *testing.T) {
	h := Header{
		ChainID:          123,
		PrevBlockHash:    common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		CrossStateRoot:   common.UINT256_EMPTY,
		BlockRoot:        common.UINT256_EMPTY,
		Timestamp:        123,
		Height:           123,
		ConsensusData:    123,
		ConsensusPayload: []byte{123},
		NextBookkeeper:   common.ADDRESS_EMPTY,
	}
	sink := common.NewZeroCopySink(nil)
	err := h.Serialization(sink)
	assert.NoError(t, err)

	buf := bytes.NewBuffer(nil)
	err = h.Serialize(buf)

	assert.NoError(t, err)
	assert.Equal(t, sink.Bytes(), buf.Bytes())

	var header1 Header
	err = header1.Deserialize(buf)
	assert.NoError(t, err)

	var header2 Header

	err = header2.Deserialization(common.NewZeroCopySource(sink.Bytes()))

	assert.NoError(t, err)

	assert.Equal(t, header1, header2)

}
