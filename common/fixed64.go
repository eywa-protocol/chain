package common

import (
	"bytes"
	"io"
	"strconv"
)

//the 64 bit fixed-point number, precise 10^-8
type Fixed64 int64

const (
	Decimal = 100000000
)

func (f *Fixed64) Serialization(sink *ZeroCopySink) {
	sink.WriteInt64(int64(*f))
}

func (f *Fixed64) Deserialization(source *ZeroCopySource) error {
	var eof bool

	t, eof := source.NextInt64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	*f = Fixed64(t)
	return nil
}

func FromDecimal(value int64) Fixed64 {
	return Fixed64(value * Decimal)
}

func (f Fixed64) GetData() int64 {
	return int64(f)
}

func (f Fixed64) String() string {
	var buffer bytes.Buffer
	value := int64(f)
	if value < 0 {
		buffer.WriteRune('-')
		value = -value
	}
	buffer.WriteString(strconv.FormatInt(value/100000000, 10))
	value %= 100000000
	if value > 0 {
		buffer.WriteRune('.')
		s := strconv.FormatInt(value, 10)
		for i := len(s); i < 8; i++ {
			buffer.WriteRune('0')
		}
		buffer.WriteString(s)
	}
	return buffer.String()
}
