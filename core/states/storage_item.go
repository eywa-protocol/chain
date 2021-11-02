/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/

package states

import (
	"bytes"
	"fmt"
	"io"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/serialization"
)

type StorageItem struct {
	StateBase
	Value []byte
}

func (this *StorageItem) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteVarBytes(w, this.Value)
	return nil
}

func (this *StorageItem) Deserialize(r io.Reader) error {
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return fmt.Errorf("[StorageItem], StateBase Deserialize failed, error:%s", err)
	}
	value, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[StorageItem], Value Deserialize failed, error:%s", err)
	}
	this.Value = value
	return nil
}

func (storageItem *StorageItem) ToArray() []byte {
	b := new(bytes.Buffer)
	storageItem.Serialize(b)
	return b.Bytes()
}

func GetValueFromRawStorageItem(raw []byte) ([]byte, error) {
	item := StorageItem{}
	err := item.Deserialize(bytes.NewBuffer(raw))
	if err != nil {
		return nil, err
	}

	return item.Value, nil
}

func GenRawStorageItem(value []byte) []byte {
	item := StorageItem{}
	item.Value = value
	return item.ToArray()
}
