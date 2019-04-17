package byteorder

import (
	"errors"
	"io"
	"math"
)

const (
	UVMAXLEN = 10
	// false stands for msb
	// true stands for lsb
	BigEndian    = ByteOrder(false)
	LittleEndian = ByteOrder(true)
)

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

// lsb means decode the first group as the right most 7 bits
//
// msb means decode the first group as the left most 7 bits
func (t ByteOrder) UVarint(b io.Reader) (uint64, error) {
	c := 0
	var ch [16]byte

	for {
		if _, e := b.Read(ch[c : c+1]); e != nil {
			return 0, e
		}

		if ch[c]&0x80 == 0 {
			break
		}

		c++
	}

	r, _, e := t.UVarintB(ch[:c+1])
	return r, e
}

// zigzag encoding
func (t ByteOrder) Varint(b io.Reader) (int64, error) {
	c := 0
	var ch [16]byte

	for {
		if _, e := b.Read(ch[c : c+1]); e != nil {
			return 0, e
		}

		if ch[c]&0x80 == 0 {
			break
		}

		c++
	}

	r, _, e := t.VarintB(ch[:c+1])
	return r, e
}

// slice version
func (t ByteOrder) UVarintB(ch []byte) (uint64, int, error) {
	r := uint64(0)
	c := 0

	if t {
		s := uint(0)
		for {
			if c == UVMAXLEN {
				return 0, 0, errors.New("overflowed uint64")
			}

			if ch[c]&0x80 == 0 {
				r |= uint64(ch[c]&0x7F) << s
				break
			}

			r |= uint64(ch[c]&0x7F) << s
			s += 7
			c++
		}
	} else {
		for {
			if c == UVMAXLEN {
				return 0, 0, errors.New("overflowed uint64")
			}

			if ch[c]&0x80 == 0 {
				r <<= 7
				r |= uint64(ch[c] & 0x7F)
				break
			}

			r <<= 7
			r |= uint64(ch[c] & 0x7F)
			c++
		}
	}

	return r, c + 1, nil
}

// slice version
func (t ByteOrder) VarintB(b []byte) (int64, int, error) {
	ux, l, e := t.UVarintB(b)

	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}

	return x, l, e
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

func (t ByteOrder) PutUVarint(out io.Writer, v uint64) error {
	var ch [16]byte

	l := t.PutUVarintB(ch[:], v)

	_, e := out.Write(ch[:l])
	if e != nil {
		return e
	}

	return nil
}

func (t ByteOrder) PutVarint(out io.Writer, v int64) error {
	var ch [16]byte

	l := t.PutVarintB(ch[:], v)

	_, e := out.Write(ch[:l])
	if e != nil {
		return e
	}

	return nil
}

func (t ByteOrder) PutUVarintB(b []byte, v uint64) int {
	if t {
		c := 0

		for v >= 0x80 {
			b[c] = byte(v&0x7F) | 0x80
			v >>= 7
			c++
		}
		b[c] = byte(v & 0x7F)

		return c + 1
	} else {
		var buf [16]byte
		c := 15

		buf[c] = byte(v & 0x7F)
		for v >= 0x80 {
			v >>= 7
			c--

			buf[c] = byte(v&0x7F) | 0x80
		}

		copy(b, buf[c:])

		return 16 - c
	}
}

// zigzag encoding
func (t ByteOrder) PutVarintB(b []byte, v int64) int {
	uv := uint64(v) << 1
	if v < 0 {
		uv = ^uv
	}

	return t.PutUVarintB(b, uv)
}

func (t ByteOrder) String() string {
	if t {
		return "LittleEndian"
	} else {
		return "BigEndian"
	}
}
