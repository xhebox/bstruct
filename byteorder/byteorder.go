package byteorder

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"math/bits"
	"reflect"
	"unsafe"
)

const (
	VMAXLEN = 10
)

var (
	BigEndian    = &bigEndian{ByteOrder: binary.BigEndian}
	LittleEndian = &littleEndian{ByteOrder: binary.LittleEndian}
	HostEndian   ByteOrder
)

func init() {
	test := 0xABCD
	if uint8(test) == 0xAB {
		HostEndian = BigEndian
	} else {
		HostEndian = LittleEndian
	}
}

// reverse buf should have a buffer, which size is a multiple of the second argument
func ReverseBuf(b []byte, size int) {
	if size == 1 {
		return
	}

	blen := len(b)

	switch size {
	case 2:
		buf := *(*[]uint16)(unsafe.Pointer(&reflect.SliceHeader{
			Data: (*reflect.SliceHeader)(unsafe.Pointer(&b)).Data,
			Len:  blen >> 1,
			Cap:  blen >> 1,
		}))

		for v, e := 0, len(buf); v < e; v++ {
			buf[v] = bits.Reverse16(buf[v])
		}
	case 4:
		buf := *(*[]uint32)(unsafe.Pointer(&reflect.SliceHeader{
			Data: (*reflect.SliceHeader)(unsafe.Pointer(&b)).Data,
			Len:  blen >> 2,
			Cap:  blen >> 2,
		}))

		for v, e := 0, len(buf); v < e; v++ {
			buf[v] = bits.Reverse32(buf[v])
		}
	case 8:
		buf := *(*[]uint64)(unsafe.Pointer(&reflect.SliceHeader{
			Data: (*reflect.SliceHeader)(unsafe.Pointer(&b)).Data,
			Len:  blen >> 3,
			Cap:  blen >> 3,
		}))

		for v, e := 0, len(buf); v < e; v++ {
			buf[v] = bits.Reverse64(buf[v])
		}
	default:
		blen1 := size / 2
		blen2 := size - 1
		for v := 0; v < blen; v += size {
			for n := 0; n < blen1; n++ {
				b[v+n], b[v+blen2-n] = b[v+blen2-n], b[v+n]
			}
		}
	}
}

func ReverseBytes(b []byte) {
	blen := len(b)

	if blen == 1 {
		return
	}

	blen1 := blen / 2
	blen2 := blen - 1

	for n := 0; n < blen1; n++ {
		b[n], b[blen2-n] = b[blen2-n], b[n]
	}
}

func Bool2Byte(r bool) byte {
	if r {
		return 1
	} else {
		return 0
	}
}

func Byte2Bool(r byte) bool {
	return r != 0
}

func Uint8(rd io.Reader) (uint8, error) {
	buf := make([]byte, 1)

	if _, e := io.ReadFull(rd, buf); e != nil {
		return 0, e
	}

	return buf[0], nil
}

func Bool(rd io.Reader) (bool, error) {
	u, e := Uint8(rd)
	return u != 0, e
}

func Int8(rd io.Reader) (int8, error) {
	u, e := Uint8(rd)
	return int8(u), e
}

func Uint16(rd io.Reader, t ByteOrder) (uint16, error) {
	buf := make([]byte, 2)

	if _, e := io.ReadFull(rd, buf); e != nil {
		return 0, e
	}

	return t.Uint16(buf), nil
}

func Int16(rd io.Reader, t ByteOrder) (int16, error) {
	u, e := Uint16(rd, t)
	return int16(u), e
}

func Uint32(rd io.Reader, t ByteOrder) (uint32, error) {
	buf := make([]byte, 4)

	if _, e := io.ReadFull(rd, buf); e != nil {
		return 0, e
	}

	return t.Uint32(buf), nil
}

func Int32(rd io.Reader, t ByteOrder) (int32, error) {
	u, e := Uint32(rd, t)
	return int32(u), e
}

func Uint64(rd io.Reader, t ByteOrder) (uint64, error) {
	buf := make([]byte, 8)

	if _, e := io.ReadFull(rd, buf); e != nil {
		return 0, e
	}

	return t.Uint64(buf), nil
}

func Int64(rd io.Reader, t ByteOrder) (int64, error) {
	u, e := Uint64(rd, t)
	return int64(u), e
}

func Float32(rd io.Reader, t ByteOrder) (float32, error) {
	u, e := Uint32(rd, t)
	return math.Float32frombits(u), e
}

func Float64(rd io.Reader, t ByteOrder) (float64, error) {
	u, e := Uint64(rd, t)
	return math.Float64frombits(u), e
}

func UVarint(b io.Reader, t ByteOrder) (uint64, error) {
	ch := make([]byte, VMAXLEN)
	c := 0

	for ; c < VMAXLEN; c++ {
		if _, e := io.ReadFull(b, ch[c:c+1]); e != nil {
			return 0, e
		}

		if ch[c]&0x80 == 0 {
			break
		}
	}

	if c == VMAXLEN && ch[c-1]&0x80 != 0 {
		return 0, errors.New("overflowed")
	}

	r, _, e := t.UVarint(ch[:])
	return r, e
}

// zigzag encoding
func Varint(rd io.Reader, t ByteOrder) (int64, error) {
	ch := make([]byte, VMAXLEN)
	c := 0

	for ; c < VMAXLEN; c++ {
		if _, e := io.ReadFull(rd, ch[c:c+1]); e != nil {
			return 0, e
		}

		if ch[c]&0x80 == 0 {
			break
		}
	}

	if c == VMAXLEN && ch[c-1]&0x80 != 0 {
		return 0, errors.New("overflowed")
	}

	r, _, e := t.Varint(ch[:])
	return r, e
}

func PutUint8(wt io.Writer, v uint8) error {
	buf := []byte{v}

	_, e := wt.Write(buf)
	if e != nil {
		return e
	}

	return nil
}

func PutBool(wt io.Writer, v bool) error {
	if v {
		return PutUint8(wt, 1)
	} else {
		return PutUint8(wt, 0)
	}
}

func PutInt8(wt io.Writer, v int8) error {
	return PutUint8(wt, uint8(v))
}

func PutUint16(wt io.Writer, t ByteOrder, v uint16) error {
	buf := make([]byte, 2)

	t.PutUint16(buf, v)

	_, e := wt.Write(buf)
	if e != nil {
		return e
	}

	return nil
}

func PutInt16(wt io.Writer, t ByteOrder, v int16) error {
	return PutUint16(wt, t, uint16(v))
}

func PutUint32(wt io.Writer, t ByteOrder, v uint32) error {
	buf := make([]byte, 4)

	t.PutUint32(buf, v)

	_, e := wt.Write(buf)
	if e != nil {
		return e
	}

	return nil
}

func PutInt32(wt io.Writer, t ByteOrder, v int32) error {
	return PutUint32(wt, t, uint32(v))
}

func PutUint64(wt io.Writer, t ByteOrder, v uint64) error {
	buf := make([]byte, 8)

	t.PutUint64(buf, v)

	_, e := wt.Write(buf)
	if e != nil {
		return e
	}

	return nil
}

func PutInt64(wt io.Writer, t ByteOrder, v int64) error {
	return PutUint64(wt, t, uint64(v))
}

func PutFloat32(wt io.Writer, t ByteOrder, v float32) error {
	return PutUint32(wt, t, math.Float32bits(v))
}

func PutFloat64(wt io.Writer, t ByteOrder, v float64) error {
	return PutUint64(wt, t, math.Float64bits(v))
}

func PutUVarint(out io.Writer, t ByteOrder, v uint64) error {
	ch := make([]byte, 16)

	l := t.PutUVarint(ch, v)

	_, e := out.Write(ch[:l])
	if e != nil {
		return e
	}

	return nil
}

func PutVarint(wt io.Writer, t ByteOrder, v int64) error {
	ch := make([]byte, 16)

	l := t.PutVarint(ch, v)

	_, e := wt.Write(ch[:l])
	if e != nil {
		return e
	}

	return nil
}

type ByteOrder interface {
	Uint16([]byte) uint16
	Uint32([]byte) uint32
	Uint64([]byte) uint64
	Int16([]byte) int16
	Int32([]byte) int32
	Int64([]byte) int64
	Float32([]byte) float32
	Float64([]byte) float64

	PutUint16([]byte, uint16) int
	PutUint32([]byte, uint32) int
	PutUint64([]byte, uint64) int
	PutInt16([]byte, int16) int
	PutInt32([]byte, int32) int
	PutInt64([]byte, int64) int
	PutFloat32([]byte, float32) int
	PutFloat64([]byte, float64) int

	UVarint([]byte) (uint64, int, error)
	PutUVarint([]byte, uint64) int

	Varint([]byte) (int64, int, error)
	PutVarint([]byte, int64) int

	String() string
}

func varint(t ByteOrder, b []byte) (int64, int, error) {
	ux, l, e := t.UVarint(b)

	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}

	return x, l, e
}

func putVarint(t ByteOrder, b []byte, v int64) int {
	uv := uint64(v) << 1
	if v < 0 {
		uv = ^uv
	}

	return t.PutUVarint(b, uv)
}

type bigEndian struct {
	binary.ByteOrder
}

func (t bigEndian) Int16(b []byte) int16 {
	return int16(t.Uint16(b))
}

func (t bigEndian) Int32(b []byte) int32 {
	return int32(t.Uint32(b))
}

func (t bigEndian) Int64(b []byte) int64 {
	return int64(t.Uint64(b))
}

func (t bigEndian) Float32(b []byte) float32 {
	return math.Float32frombits(t.Uint32(b))
}

func (t bigEndian) Float64(b []byte) float64 {
	return math.Float64frombits(t.Uint64(b))
}

func (t bigEndian) PutUint16(b []byte, h uint16) int {
	t.ByteOrder.PutUint16(b, h)
	return 2
}

func (t bigEndian) PutUint32(b []byte, h uint32) int {
	t.ByteOrder.PutUint32(b, h)
	return 4
}

func (t bigEndian) PutUint64(b []byte, h uint64) int {
	t.ByteOrder.PutUint64(b, h)
	return 8
}

func (t bigEndian) PutInt16(b []byte, h int16) int {
	return t.PutUint16(b, uint16(h))
}

func (t bigEndian) PutInt32(b []byte, h int32) int {
	return t.PutUint32(b, uint32(h))
}

func (t bigEndian) PutInt64(b []byte, h int64) int {
	return t.PutUint64(b, uint64(h))
}

func (t bigEndian) PutFloat32(b []byte, h float32) int {
	return t.PutUint32(b, math.Float32bits(h))
}

func (t bigEndian) PutFloat64(b []byte, h float64) int {
	return t.PutUint64(b, math.Float64bits(h))
}

// lsb means decode the first group as the right most 7 bits
//
// msb means decode the first group as the left most 7 bits
func (t bigEndian) UVarint(b []byte) (uint64, int, error) {
	r := uint64(0)

	c := 0

	for ; b[c]&0x80 != 0; c++ {
		r = r<<7 | uint64(b[c]&0x7F)
	}

	r = r<<7 | uint64(b[c]&0x7F)

	return r, c + 1, nil
}

func (t bigEndian) PutUVarint(b []byte, v uint64) int {
	c := 0

	for ; v != 0; c++ {
		b[c] = byte(v&0x7F) | 0x80
		v >>= 7
	}

	ReverseBytes(b[:c])

	b[c-1] &= 0x7F
	return c
}

func (t bigEndian) Varint(b []byte) (int64, int, error) {
	return varint(t, b)
}

func (t bigEndian) PutVarint(b []byte, v int64) int {
	return putVarint(t, b, v)
}

type littleEndian struct {
	binary.ByteOrder
}

func (t littleEndian) Int16(b []byte) int16 {
	return int16(t.Uint16(b))
}

func (t littleEndian) Int32(b []byte) int32 {
	return int32(t.Uint32(b))
}

func (t littleEndian) Int64(b []byte) int64 {
	return int64(t.Uint64(b))
}

func (t littleEndian) Float32(b []byte) float32 {
	return math.Float32frombits(t.Uint32(b))
}

func (t littleEndian) Float64(b []byte) float64 {
	return math.Float64frombits(t.Uint64(b))
}

func (t littleEndian) PutUint16(b []byte, h uint16) int {
	t.ByteOrder.PutUint16(b, h)
	return 2
}

func (t littleEndian) PutUint32(b []byte, h uint32) int {
	t.ByteOrder.PutUint32(b, h)
	return 4
}

func (t littleEndian) PutUint64(b []byte, h uint64) int {
	t.ByteOrder.PutUint64(b, h)
	return 8
}

func (t littleEndian) PutInt16(b []byte, h int16) int {
	return t.PutUint16(b, uint16(h))
}

func (t littleEndian) PutInt32(b []byte, h int32) int {
	return t.PutUint32(b, uint32(h))
}

func (t littleEndian) PutInt64(b []byte, h int64) int {
	return t.PutUint64(b, uint64(h))
}

func (t littleEndian) PutFloat32(b []byte, h float32) int {
	return t.PutUint32(b, math.Float32bits(h))
}

func (t littleEndian) PutFloat64(b []byte, h float64) int {
	return t.PutUint64(b, math.Float64bits(h))
}

func (t littleEndian) UVarint(b []byte) (uint64, int, error) {
	r := uint64(0)

	c := 0

	for ; b[c]&0x80 != 0; c++ {
		r |= uint64(b[c]&0x7F) << uint(c*7)
	}

	r |= uint64(b[c]&0x7F) << uint(c*7)

	return r, c + 1, nil
}

func (t littleEndian) PutUVarint(b []byte, v uint64) int {
	c := 0

	for ; v != 0; c++ {
		b[c] = byte(v&0x7F) | 0x80
		v >>= 7
	}

	b[c-1] &= 0x7F
	return c
}

func (t littleEndian) Varint(b []byte) (int64, int, error) {
	return varint(t, b)
}

func (t littleEndian) PutVarint(b []byte, v int64) int {
	return putVarint(t, b, v)
}
