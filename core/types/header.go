package types

import (
	"crypto/sha256"
	"encoding/binary"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/eywa-protocol/bls-crypto/bls"
	"github.com/eywa-protocol/chain/common"
)

const (
	BlockHeaderV1 = 1
	BlockHeaderV2 = 2
)

type Header struct {
	ChainID          uint64
	PrevBlockHash    common.Uint256
	EpochBlockHash   common.Uint256
	TransactionsRoot common.Uint256
	SourceHeight     uint64
	Height           uint64
	Signature        bls.Multisig
	TimeStamp        time.Time

	hash *common.Uint256
}

const BLOCK_SIZE = 124

func (bd *Header) Serialization(sink *common.ZeroCopySink) error {
	bd.serializationUnsigned(sink, BlockHeaderV2)
	sink.WriteVarBytes(bd.Signature.PartSignature.Marshal())
	sink.WriteVarBytes(bd.Signature.PartPublicKey.Marshal())
	sink.WriteVarBytes(bls.MarshalBitmask(bd.Signature.PartMask))
	return nil
}

func (bd *Header) serializationUnsigned(sink *common.ZeroCopySink, version int8) {
	sink.WriteUint64(bd.ChainID)
	sink.WriteBytes(bd.PrevBlockHash[:])
	sink.WriteBytes(bd.EpochBlockHash[:])
	sink.WriteBytes(bd.TransactionsRoot[:])
	sink.WriteUint64(bd.SourceHeight)
	sink.WriteUint64(bd.Height)
	if version > BlockHeaderV1 {
		timeBin, err := bd.TimeStamp.MarshalBinary()
		if err != nil {
			logrus.Errorf("error marhsal block.timestamp: %s", err)
		} else {
			sink.WriteBytes(timeBin)
		}
	}
}

func (bd *Header) Serialize(w io.Writer) error {
	sink := common.NewZeroCopySink(nil)
	err := bd.Serialization(sink)
	if err != nil {
		return err
	}
	_, err = w.Write(sink.Bytes())
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
	err := bd.tryDeserializationVersion(source, BlockHeaderV2)
	if err != nil {
		// Try version 1 and repeat deserialize
		logrus.Warnf("try block version to 1 for deserialize (%s)", err)
		source.Reset()
		err = bd.tryDeserializationVersion(source, BlockHeaderV1)
		if err != nil {
			return errors.Wrap(err, "error deserialize unsigned part")
		}
	}

	return nil
}

func (bd *Header) tryDeserializationVersion(source *common.ZeroCopySource, version int8) error {
	err := bd.deserializationUnsigned(source, version)
	if err != nil {
		return errors.Wrap(err, "error deserialize unsigned part")
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

func (bd *Header) deserializationUnsigned(source *common.ZeroCopySource, version int8) error {
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

	if version == BlockHeaderV2 {
		// Deserialize timestamp only if header version is 2
		var err error
		ts := &(bd.TimeStamp)
		switch source.OffBytes()[0] {
		case 0x01:
			err = ts.UnmarshalBinary(source.OffBytes()[:15])
		case 0x02:
			err = ts.UnmarshalBinary(source.OffBytes()[:16])
		default:
			err = errors.New("bad version of timestamp binary")
		}
		if err != nil {
			logrus.Errorf("can not unmarshal timestamp: %s", err)
			return errors.Wrap(err, "can not unmarshal timestamp")
		}
		// If timestamp no unmarshal errors - skip timestamp bytes
		switch source.OffBytes()[0] {
		case 0x01:
			// timeBinaryVersionV1
			source.Skip(15)
		case 0x02:
			// timeBinaryVersionV2
			source.Skip(16)
		}
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
	data = append(data, rawUint64(uint64(bd.TimeStamp.Unix()))...)
	return data
}

func (bd *Header) Hash() *common.Uint256 {
	return bd.hash
}

func (bd *Header) сalculateHash() {
	hash := common.Uint256(sha256.Sum256(bd.RawData()))
	bd.hash = &hash
}
