/*
* Copyright 2021 by EYWA chain <blockchain@digiu.ai>
*/
package common

type Serializable interface {
	Serialization(sink *ZeroCopySink)
}

func SerializeToBytes(values ...Serializable) []byte {
	sink := NewZeroCopySink(nil)
	for _, val := range values {
		val.Serialization(sink)
	}

	return sink.Bytes()
}
