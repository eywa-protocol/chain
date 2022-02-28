package types

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"math/big"

	"github.com/eywa-protocol/bls-crypto/bls"

	"github.com/eywa-protocol/chain/common"
)

// CryptoProof is a BLS multisignature proof that anyone may provide to
// convince the verifier that the listed nodes has signed the message
// TODO: move this and session.CryproProof to bls repo.
type CryptoProof struct {
	PartSignature bls.Signature // Aggregated partial signature collected by this node
	PartPublicKey bls.PublicKey // Aggregated partial public key collected by this node
	SigMask       big.Int       // Bitmask of nodes that the correct signature is received from
}

type Header struct {
	ChainID          uint64
	PrevBlockHash    common.Uint256
	EpochBlockHash   common.Uint256
	TransactionsRoot common.Uint256
	SourceHeight     uint64
	Height           uint64
	Signature        CryptoProof
	hash             *common.Uint256
}

const BLOCK_SIZE = 124

func (bd *Header) Serialization(sink *common.ZeroCopySink) error {
	bd.serializationUnsigned(sink)
	sink.WriteVarBytes(bd.Signature.PartSignature.Marshal())
	sink.WriteVarBytes(bd.Signature.PartPublicKey.Marshal())
	sink.WriteString(bd.Signature.SigMask.Text(16))
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

	sigMask, eof := source.NextString()
	if eof {
		return errors.New("[Header] deserialize sigMask error")
	}
	err = json.Unmarshal([]byte(sigMask), &bd.Signature.SigMask)
	if err != nil {
		return errors.New("[Header] unmarshal sigMask error")
	}

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

func (bd *Header) Hash() common.Uint256 {
	if bd.hash != nil {
		return *bd.hash
	}
	sink := common.NewZeroCopySink(nil)
	bd.serializationUnsigned(sink)
	temp := sha256.Sum256(sink.Bytes())
	hash := common.Uint256(sha256.Sum256(temp[:]))

	bd.hash = &hash
	return hash
}

func (bd *Header) GetMessage() []byte {
	sink := common.NewZeroCopySink(nil)
	bd.serializationUnsigned(sink)
	return sink.Bytes()
}

func (bd *Header) ToArray() []byte {
	sink := common.NewZeroCopySink(nil)
	bd.Serialization(sink)
	return sink.Bytes()
}
