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
	return t.Uint8(b) != 0
}

func (t ByteOrder) FBool(rd io.Reader) (bool, error) {
	u, e := t.FUint8(rd)
	return u != 0, e
}

func (t ByteOrder) Uint8(b []byte) uint8 {
	return b[0]
}

func (t ByteOrder) FUint8(rd io.Reader) (uint8, error) {
	buf := []byte{0}

	_, e := rd.Read(buf)
	if e != nil {
		return 0, e
	}

	return t.Uint8(buf), nil
}

func (t ByteOrder) Int8(b []byte) int8 {
	return int8(t.Uint8(b))
}

func (t ByteOrder) FInt8(rd io.Reader) (int8, error) {
	u, e := t.FUint8(rd)
	return int8(u), e
}

func (t ByteOrder) Uint16(b []byte) uint16 {
	_ = b[1]
	if t {
		return uint16(b[0]) | uint16(b[1])<<8
	} else {
		return uint16(b[1]) | uint16(b[0])<<8
	}
}

func (t ByteOrder) FUint16(rd io.Reader) (uint16, error) {
	buf := []byte{0, 0}

	_, e := rd.Read(buf)
	if e != nil {
		return 0, e
	}

	return t.Uint16(buf), nil
}

func (t ByteOrder) Int16(b []byte) int16 {
	return int16(t.Uint16(b))
}

func (t ByteOrder) FInt16(rd io.Reader) (int16, error) {
	u, e := t.FUint16(rd)
	return int16(u), e
}

func (t ByteOrder) Uint32(b []byte) uint32 {
	_ = b[3]
	if t {
		return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	} else {
		return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
	}
}

func (t ByteOrder) FUint32(rd io.Reader) (uint32, error) {
	buf := []byte{0, 0, 0, 0}

	_, e := rd.Read(buf)
	if e != nil {
		return 0, e
	}

	return t.Uint32(buf), nil
}

func (t ByteOrder) Int32(b []byte) int32 {
	return int32(t.Uint32(b))
}

func (t ByteOrder) FInt32(rd io.Reader) (int32, error) {
	u, e := t.FUint32(rd)
	return int32(u), e
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

func (t ByteOrder) FUint64(rd io.Reader) (uint64, error) {
	buf := []byte{0, 0, 0, 0, 0, 0, 0, 0}

	_, e := rd.Read(buf)
	if e != nil {
		return 0, e
	}

	return t.Uint64(buf), nil
}

func (t ByteOrder) Int64(b []byte) int64 {
	return int64(t.Uint64(b))
}

func (t ByteOrder) FInt64(rd io.Reader) (int64, error) {
	u, e := t.FUint64(rd)
	return int64(u), e
}

func (t ByteOrder) Float32(b []byte) float32 {
	return math.Float32frombits(t.Uint32(b))
}

func (t ByteOrder) FFloat32(rd io.Reader) (float32, error) {
	u, e := t.FUint32(rd)
	return math.Float32frombits(u), e
}

func (t ByteOrder) Float64(b []byte) float64 {
	return math.Float64frombits(t.Uint64(b))
}

func (t ByteOrder) FFloat64(rd io.Reader) (float64, error) {
	u, e := t.FUint64(rd)
	return math.Float64frombits(u), e
}

// lsb means decode the first group as the right most 7 bits
//
// msb means decode the first group as the left most 7 bits
func (t ByteOrder) UVarint(ch []byte) (uint64, int, error) {
	r := uint64(0)

	c := 0

	for ; c < UVMAXLEN && ch[c]&0x80 == 0; c++ {
		if t {
			r |= uint64(ch[c]&0x7F) << uint(c*7)
		} else {
			r = r<<7 | uint64(ch[c]&0x7F)
		}
	}

	if c == UVMAXLEN {
		return 0, 0, errors.New("overflowed uint64")
	}

	return r, c + 1, nil
}

func (t ByteOrder) FUVarint(b io.Reader) (uint64, error) {
	var ch [UVMAXLEN]byte

	for c := 0; c < UVMAXLEN && ch[c]&0x80 != 0; c++ {
		if _, e := b.Read(ch[c : c+1]); e != nil {
			return 0, e
		}
	}

	r, _, e := t.UVarint(ch[:])
	return r, e
}

func (t ByteOrder) Varint(b []byte) (int64, int, error) {
	ux, l, e := t.UVarint(b)

	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}

	return x, l, e
}

// zigzag encoding
func (t ByteOrder) FVarint(b io.Reader) (int64, error) {
	var ch [UVMAXLEN]byte

	for c := 0; c < UVMAXLEN && ch[c]&0x80 != 0; c++ {
		if _, e := b.Read(ch[c : c+1]); e != nil {
			return 0, e
		}
	}

	r, _, e := t.Varint(ch[:])
	return r, e
}

func (t ByteOrder) PutBool(b []byte, v bool) {
	if v {
		t.PutUint8(b, 1)
	} else {
		t.PutUint8(b, 0)
	}
}

func (t ByteOrder) FputBool(wt io.Writer, v bool) error {
	if v {
		return t.FputUint8(wt, 1)
	} else {
		return t.FputUint8(wt, 0)
	}
}

func (t ByteOrder) PutUint8(b []byte, v uint8) {
	b[0] = v
}

func (t ByteOrder) FputUint8(wt io.Writer, v uint8) error {
	buf := []byte{0}

	t.PutUint8(buf, v)

	_, e := wt.Write(buf)
	if e != nil {
		return e
	}

	return nil
}

func (t ByteOrder) PutInt8(b []byte, v int8) {
	t.PutUint8(b, uint8(v))
}

func (t ByteOrder) FputInt8(wt io.Writer, v int8) error {
	return t.FputUint8(wt, uint8(v))
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

func (t ByteOrder) FputUint16(wt io.Writer, v uint16) error {
	buf := []byte{0, 0}

	t.PutUint16(buf, v)

	_, e := wt.Write(buf)
	if e != nil {
		return e
	}

	return nil
}

func (t ByteOrder) PutInt16(b []byte, v int16) {
	t.PutUint16(b, uint16(v))
}

func (t ByteOrder) FputInt16(wt io.Writer, v int16) error {
	return t.FputUint16(wt, uint16(v))
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

func (t ByteOrder) FputUint32(wt io.Writer, v uint32) error {
	buf := []byte{0, 0, 0, 0}

	t.PutUint32(buf, v)

	_, e := wt.Write(buf)
	if e != nil {
		return e
	}

	return nil
}

func (t ByteOrder) PutInt32(b []byte, v int32) {
	t.PutUint32(b, uint32(v))
}

func (t ByteOrder) FputInt32(wt io.Writer, v int32) error {
	return t.FputUint32(wt, uint32(v))
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

func (t ByteOrder) FputUint64(wt io.Writer, v uint64) error {
	buf := []byte{0, 0, 0, 0, 0, 0, 0, 0}

	t.PutUint64(buf, v)

	_, e := wt.Write(buf)
	if e != nil {
		return e
	}

	return nil
}

func (t ByteOrder) PutInt64(b []byte, v int64) {
	t.PutUint64(b, uint64(v))
}

func (t ByteOrder) FputInt64(wt io.Writer, v int64) error {
	return t.FputUint64(wt, uint64(v))
}

func (t ByteOrder) PutFloat32(b []byte, v float32) {
	t.PutUint32(b, math.Float32bits(v))
}

func (t ByteOrder) FputFloat32(wt io.Writer, v float32) error {
	return t.FputUint32(wt, math.Float32bits(v))
}

func (t ByteOrder) PutFloat64(b []byte, v float64) {
	t.PutUint64(b, math.Float64bits(v))
}

func (t ByteOrder) FputFloat64(wt io.Writer, v float64) error {
	return t.FputUint64(wt, math.Float64bits(v))
}

func (t ByteOrder) PutUVarint(b []byte, v uint64) int {
	c := 0
	for ; v != 0; c++ {
		b[c] = byte(v&0x7F) | 0x80
		v >>= 7
	}

	if t {
		b[c-1] &= 0x7F
		return c
	}

	for i := 0; 2*i < c; i++ {
		tmp := b[c-i]
		b[c-i] = b[i]
		b[i] = tmp
	}

	return c
}

func (t ByteOrder) FputUVarint(out io.Writer, v uint64) error {
	var ch [16]byte

	l := t.PutUVarint(ch[:], v)

	_, e := out.Write(ch[:l])
	if e != nil {
		return e
	}

	return nil
}

func (t ByteOrder) PutVarint(b []byte, v int64) int {
	uv := uint64(v) << 1
	if v < 0 {
		uv = ^uv
	}

	return t.PutUVarint(b, uv)
}

func (t ByteOrder) FputVarint(out io.Writer, v int64) error {
	var ch [16]byte

	l := t.PutVarint(ch[:], v)

	_, e := out.Write(ch[:l])
	if e != nil {
		return e
	}

	return nil
}

func (t ByteOrder) String() string {
	if t {
		return "LittleEndian"
	} else {
		return "BigEndian"
	}
}
