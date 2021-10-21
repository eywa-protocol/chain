/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */

package states

import (
	"github.com/eywa-protocol/bls-crypto/bls"
	"io"

	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common/serialization"
)

type VoteState struct {
	StateBase
	PublicKeys []bls.PublicKey
	Count      common.Fixed64
}

func (this *VoteState) Serialize(w io.Writer) error {
	err := this.StateBase.Serialize(w)
	if err != nil {
		return err
	}
	err = serialization.WriteUint32(w, uint32(len(this.PublicKeys)))
	if err != nil {
		return err
	}
	for _, v := range this.PublicKeys {
		buf := v.Marshal()
		err := serialization.WriteVarBytes(w, buf)
		if err != nil {
			return err
		}
	}

	return serialization.WriteUint64(w, uint64(this.Count))
}

func (this *VoteState) Deserialize(r io.Reader) error {
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return err
	}
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	for i := 0; i < int(n); i++ {
		buf, err := serialization.ReadVarBytes(r)
		if err != nil {
			return err
		}
		pk, err := bls.UnmarshalPublicKey(buf)
		if err != nil {
			return err
		}
		this.PublicKeys = append(this.PublicKeys, pk)
	}
	c, err := serialization.ReadUint64(r)
	if err != nil {
		return err
	}
	this.Count = common.Fixed64(int64(c))
	return nil
}
