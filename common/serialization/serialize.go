/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/

package serialization

import (
	"bytes"
	"encoding/binary"
	"errors"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-overhead-chain/common"
	"io"
	"math"
)

var ErrRange = errors.New("value out of range")
var ErrEof = errors.New("got EOF, can not get the next byte")

// SerializableData describe the data need be serialized.
type SerializableData interface {
	// Write data to writer
	Serialize(w io.Writer) error

	// read data to reader
	Deserialize(r io.Reader) error
}


func WriteVarUint(writer io.Writer, value uint64) error {
	var buf [9]byte
	var len = 0
	if value < 0xFD {
		buf[0] = uint8(value)
		len = 1
	} else if value <= 0xFFFF {
		buf[0] = 0xFD
		binary.LittleEndian.PutUint16(buf[1:], uint16(value))
		len = 3
	} else if value <= 0xFFFFFFFF {
		buf[0] = 0xFE
		binary.LittleEndian.PutUint32(buf[1:], uint32(value))
		len = 5
	} else {
		buf[0] = 0xFF
		binary.LittleEndian.PutUint64(buf[1:], uint64(value))
		len = 9
	}
	_, err := writer.Write(buf[:len])
	return err
}

func ReadVarUint(reader io.Reader, maxint uint64) (uint64, error) {
	var res uint64
	if maxint == 0x00 {
		maxint = math.MaxUint64
	}
	var fb [9]byte
	_, err := io.ReadFull(reader, fb[:1])
	if err != nil {
		return 0, err
	}

	if fb[0] == byte(0xfd) {
		_, err := io.ReadFull(reader, fb[1:3])
		if err != nil {
			return 0, err
		}
		res = uint64(binary.LittleEndian.Uint16(fb[1:3]))
	} else if fb[0] == byte(0xfe) {
		_, err := io.ReadFull(reader, fb[1:5])
		if err != nil {
			return 0, err
		}
		res = uint64(binary.LittleEndian.Uint32(fb[1:5]))
	} else if fb[0] == byte(0xff) {
		_, err := io.ReadFull(reader, fb[1:9])
		if err != nil {
			return 0, err
		}
		res = uint64(binary.LittleEndian.Uint64(fb[1:9]))
	} else {
		res = uint64(fb[0])
	}
	if res > maxint {
		return 0, ErrRange
	}
	return res, nil
}

func WriteVarBytes(writer io.Writer, value []byte) error {
	err := WriteVarUint(writer, uint64(len(value)))
	if err != nil {
		return err
	}
	_, err = writer.Write(value)
	return err
}

func WriteBytes(writer io.Writer, value []byte) error {
	_, err := writer.Write(value)
	return err
}

func WriteString(writer io.Writer, value string) error {
	return WriteVarBytes(writer, []byte(value))
}

func ReadVarBytes(reader io.Reader) ([]byte, error) {
	val, err := ReadVarUint(reader, 0)
	if err != nil {
		return nil, err
	}
	str, err := byteXReader(reader, val)
	if err != nil {
		return nil, err
	}
	return str, nil
}

func ReadHash(reader io.Reader) (common.Uint256, error) {
	val, err := byteXReader(reader, common.UINT256_SIZE)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return common.Uint256ParseFromBytes(val)
}

func ReadAddress(reader io.Reader) (common.Address, error) {
	val, err := byteXReader(reader, common.ADDR_LEN)
	if err != nil {
		return common.ADDRESS_EMPTY, err
	}
	return common.AddressParseFromBytes(val)
}

func ReadString(reader io.Reader) (string, error) {
	val, err := ReadVarBytes(reader)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func GetVarUintSize(value uint64) int {
	if value < 0xfd {
		return 1
	} else if value <= 0xffff {
		return 3
	} else if value <= 0xFFFFFFFF {
		return 5
	} else {
		return 9
	}
}

func ReadBytes(reader io.Reader, length uint64) ([]byte, error) {
	str, err := byteXReader(reader, length)
	if err != nil {
		return nil, err
	}
	return str, nil
}

func ReadUint8(reader io.Reader) (uint8, error) {
	var p [1]byte
	_, err := io.ReadFull(reader, p[:])
	if err != nil {
		return 0, ErrEof
	}
	return uint8(p[0]), nil
}

func ReadUint16(reader io.Reader) (uint16, error) {
	var p [2]byte
	_, err := io.ReadFull(reader, p[:])
	if err != nil {
		return 0, ErrEof
	}
	return binary.LittleEndian.Uint16(p[:]), nil
}

func ReadUint32(reader io.Reader) (uint32, error) {
	var p [4]byte
	_, err := io.ReadFull(reader, p[:])
	if err != nil {
		return 0, ErrEof
	}
	return binary.LittleEndian.Uint32(p[:]), nil
}

func ReadUint64(reader io.Reader) (uint64, error) {
	var p [8]byte
	_, err := io.ReadFull(reader, p[:])
	if err != nil {
		return 0, ErrEof
	}
	return binary.LittleEndian.Uint64(p[:]), nil
}

func WriteUint8(writer io.Writer, val uint8) error {
	var p [1]byte
	p[0] = byte(val)
	_, err := writer.Write(p[:])
	return err
}

func WriteUint16(writer io.Writer, val uint16) error {
	var p [2]byte
	binary.LittleEndian.PutUint16(p[:], val)
	_, err := writer.Write(p[:])
	return err
}

func WriteUint32(writer io.Writer, val uint32) error {
	var p [4]byte
	binary.LittleEndian.PutUint32(p[:], val)
	_, err := writer.Write(p[:])
	return err
}

func WriteUint64(writer io.Writer, val uint64) error {
	var p [8]byte
	binary.LittleEndian.PutUint64(p[:], val)
	_, err := writer.Write(p[:])
	return err
}

func ToArray(data SerializableData) []byte {
	buf := new(bytes.Buffer)
	data.Serialize(buf)
	return buf.Bytes()
}

//**************************************************************************
//**    internal func                                                    ***
//**************************************************************************
//** 2.byteXReader: read x byte and return []byte.
//** 3.byteToUint8: change byte -> uint8 and return.
//**************************************************************************

func byteXReader(reader io.Reader, x uint64) ([]byte, error) {
	if x == 0 {
		return nil, nil
	}
	//fast path to avoid buffer reallocation
	if x < 2*1024*1024 {
		p := make([]byte, x)
		_, err := io.ReadFull(reader, p)
		if err != nil {
			return nil, err
		}
		return p, nil
	}

	// normal path to avoid attack
	limited := io.LimitReader(reader, int64(x))
	buf := &bytes.Buffer{}
	n, _ := buf.ReadFrom(limited)
	if n == int64(x) {
		return buf.Bytes(), nil
	}
	return nil, ErrEof
}

func WriteBool(writer io.Writer, val bool) error {
	err := binary.Write(writer, binary.LittleEndian, val)
	return err
}

func ReadBool(reader io.Reader) (bool, error) {
	var x bool
	err := binary.Read(reader, binary.LittleEndian, &x)
	return x, err
}

func WriteByte(writer io.Writer, val byte) error {
	_, err := writer.Write([]byte{val})
	if err != nil {
		return err
	}
	return nil
}

func ReadByte(reader io.Reader) (byte, error) {
	b, err := byteXReader(reader, 1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}
