package types

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"

	"github.com/eywa-protocol/bls-crypto/bls"

	"github.com/eywa-protocol/chain/common"
)

type Header struct {
	ChainID          uint64
	PrevBlockHash    common.Uint256
	EpochBlockHash   common.Uint256
	TransactionsRoot common.Uint256
	SourceHeight     uint64
	Height           uint64
	Signature        bls.Multisig
	hash             *common.Uint256
}

const BLOCK_SIZE = 124

func (bd *Header) Serialization(sink *common.ZeroCopySink) error {
	bd.serializationUnsigned(sink)
	sink.WriteVarBytes(bd.Signature.PartSignature.Marshal())
	sink.WriteVarBytes(bd.Signature.PartPublicKey.Marshal())
	sink.WriteVarBytes(bls.MarshalBitmask(bd.Signature.PartMask))
	return nil
}

func (bd *Header) serializationUnsigned(sink *common.ZeroCopySink) {
	sink.WriteUint64(bd.ChainID)
	sink.WriteBytes(bd.PrevBlockHash[:])
	sink.WriteBytes(bd.EpochBlockHash[:])
	sink.WriteBytes(bd.TransactionsRoot[:])
	sink.WriteUint64(bd.SourceHeight)
	sink.WriteUint64(bd.Height)
}

func (bd *Header) Serialize(w io.Writer) error {
	sink := common.NewZeroCopySink(nil)
	bd.Serialization(sink)
	_, err := w.Write(sink.Bytes())
	return err
}

func HeaderFromRawBytes(raw []byte) (*Header, error) {
	source := common.NewZeroCopySource(raw)
	header := &Header{}
	err := header.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return header, nil

}

func (bd *Header) Deserialization(source *common.ZeroCopySource) error {
	err := bd.deserializationUnsigned(source)
	if err != nil {
		return err
	}

	partSig, eof := source.NextVarBytes()
	if eof {
		return errors.New("[Header] deserialize partSig error")
	}
	bd.Signature.PartSignature, err = bls.UnmarshalSignature(partSig)
	if err != nil {
		return errors.New("[Header] unmarshal partSig error")
	}

	partPub, eof := source.NextVarBytes()
	if eof {
		return errors.New("[Header] deserialize partPub error")
	}
	bd.Signature.PartPublicKey, err = bls.UnmarshalPublicKey(partPub)
	if err != nil {
		return errors.New("[Header] unmarshal partPub error")
	}

	sigMask, eof := source.NextVarBytes()
	if eof {
		return errors.New("[Header] deserialize sigMask error")
	}
	bd.Signature.PartMask = bls.UnmarshalBitmask(sigMask)

	return nil
}

func (bd *Header) deserializationUnsigned(source *common.ZeroCopySource) error {
	var eof bool
	bd.ChainID, eof = source.NextUint64()
	if eof {
		return errors.New("[Header] read chainID error")
	}
	bd.PrevBlockHash, eof = source.NextHash()
	if eof {
		return errors.New("[Header] read prevBlockHash error")
	}
	bd.EpochBlockHash, eof = source.NextHash()
	if eof {
		return errors.New("[Header] read epochBlockHash error")
	}
	bd.TransactionsRoot, eof = source.NextHash()
	if eof {
		return errors.New("[Header] read transactionsRoot error")
	}
	bd.SourceHeight, eof = source.NextUint64()
	if eof {
		return errors.New("[Header] read sourceHeight error")
	}
	bd.Height, eof = source.NextUint64()
	if eof {
		return errors.New("[Header] read height error")
	}
	return nil
}

func rawUint64(val uint64) []byte {
	raw := make([]byte, 8)
	binary.BigEndian.PutUint64(raw, val)
	return raw
}

func (bd *Header) RawData() []byte {
	var data []byte
	data = append(data, rawUint64(bd.ChainID)...)
	data = append(data, bd.PrevBlockHash.ToArray()...)
	data = append(data, bd.EpochBlockHash.ToArray()...)
	data = append(data, bd.TransactionsRoot.ToArray()...)
	data = append(data, rawUint64(bd.SourceHeight)...)
	data = append(data, rawUint64(bd.Height)...)
	return data
}

func (bd *Header) Hash() *common.Uint256 {
	return bd.hash
}

func (bd *Header) CalculateHash() {
	hash := common.Uint256(sha256.Sum256(bd.RawData()))
	bd.hash = &hash
}
