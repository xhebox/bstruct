package byteorder

import "math"

// like the official one, with some enchancements, but without interface{}
type ByteOrder bool

func (t ByteOrder) Bool(b []byte) bool {
	return b[0] != 0
}

func (t ByteOrder) Uint8(b []byte) uint8 {
	return b[0]
}

func (t ByteOrder) Int8(b []byte) int8 {
	return int8(b[0])
}

func (t ByteOrder) Uint16(b []byte) uint16 {
	_ = b[1]
	if t {
		return uint16(b[0]) | uint16(b[1])<<8
	} else {
		return uint16(b[1]) | uint16(b[0])<<8
	}
}

func (t ByteOrder) Int16(b []byte) int16 {
	return int16(t.Uint16(b))
}

func (t ByteOrder) Uint32(b []byte) uint32 {
	_ = b[3]
	if t {
		return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	} else {
		return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
	}
}

func (t ByteOrder) Int32(b []byte) int32 {
	return int32(t.Uint32(b))
}

func (t ByteOrder) Uint64(b []byte) uint64 {
	_ = b[7]
	if t {
		return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
			uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
	} else {
		return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
			uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
	}
}

func (t ByteOrder) Int64(b []byte) int64 {
	return int64(t.Uint64(b))
}

func (t ByteOrder) Float32(b []byte) float32 {
	return math.Float32frombits(t.Uint32(b))
}

func (t ByteOrder) Float64(b []byte) float64 {
	return math.Float64frombits(t.Uint64(b))
}

func (t ByteOrder) PutBool(b []byte, v bool) {
	if v {
		b[0] = 1
	} else {
		b[0] = 0
	}
}

func (t ByteOrder) PutUint8(b []byte, v uint8) {
	b[0] = v
}

func (t ByteOrder) PutInt8(b []byte, v int8) {
	t.PutUint8(b, uint8(v))
}

func (t ByteOrder) PutUint16(b []byte, v uint16) {
	_ = b[1]
	if t {
		b[0] = byte(v)
		b[1] = byte(v >> 8)
	} else {
		b[0] = byte(v >> 8)
		b[1] = byte(v)
	}
}

func (t ByteOrder) PutInt16(b []byte, v int16) {
	t.PutUint16(b, uint16(v))
}

func (t ByteOrder) PutUint32(b []byte, v uint32) {
	_ = b[3]
	if t {
		b[0] = byte(v)
		b[1] = byte(v >> 8)
		b[2] = byte(v >> 16)
		b[3] = byte(v >> 24)
	} else {
		b[0] = byte(v >> 24)
		b[1] = byte(v >> 16)
		b[2] = byte(v >> 8)
		b[3] = byte(v)
	}
}

func (t ByteOrder) PutInt32(b []byte, v int32) {
	t.PutUint32(b, uint32(v))
}

func (t ByteOrder) PutUint64(b []byte, v uint64) {
	_ = b[7]
	if t {
		b[0] = byte(v)
		b[1] = byte(v >> 8)
		b[2] = byte(v >> 16)
		b[3] = byte(v >> 24)
		b[4] = byte(v >> 32)
		b[5] = byte(v >> 40)
		b[6] = byte(v >> 48)
		b[7] = byte(v >> 56)
	} else {
		b[0] = byte(v >> 56)
		b[1] = byte(v >> 48)
		b[2] = byte(v >> 40)
		b[3] = byte(v >> 32)
		b[4] = byte(v >> 24)
		b[5] = byte(v >> 16)
		b[6] = byte(v >> 8)
		b[7] = byte(v)
	}
}

func (t ByteOrder) PutInt64(b []byte, v int64) {
	t.PutUint64(b, uint64(v))
}

func (t ByteOrder) PutFloat32(b []byte, v float32) {
	t.PutUint32(b, math.Float32bits(v))
}

func (t ByteOrder) PutFloat64(b []byte, v float64) {
	t.PutUint64(b, math.Float64bits(v))
}