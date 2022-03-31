package test

import (
	"fmt"
	"testing"

	"github.com/eywa-protocol/chain/common"
	"github.com/stretchr/testify/assert"
)

func TestHexAndBytesTransfer(t *testing.T) {
	testBytes := []byte("10, 11, 12, 13, 14, 15, 16, 17, 18, 19")
	stringAfterTrans := common.ToHexString(testBytes)
	bytesAfterTrans, err := common.HexToBytes(stringAfterTrans)
	assert.Nil(t, err)
	assert.Equal(t, testBytes, bytesAfterTrans)
}

func TestGetNonce(t *testing.T) {
	nonce1 := common.GetNonce()
	nonce2 := common.GetNonce()
	assert.NotEqual(t, nonce1, nonce2)
}

// TODO adopt  testFileExisted test

func TestFileExisted(t *testing.T) {
	assert.True(t, common.FileExisted("common_test.go"))
	assert.True(t, common.FileExisted("../common.go"))
	assert.False(t, common.FileExisted("../log/log.og"))
}

func TestBase58(t *testing.T) {
	addr := common.ADDRESS_EMPTY
	fmt.Println("emtpy addr:", addr.ToBase58())
}
