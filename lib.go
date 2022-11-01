package bstruct

import (
	"encoding/binary"
	"unsafe"
)

//go:noescape
//go:linkname memmove runtime.memmove
func memmove(to, from unsafe.Pointer, n uintptr)

type Reader struct {
	data []byte
	pos  int
}

func NewReader(data []byte) *Reader {
	return &Reader{data: data}
}

func (r *Reader) Data() []byte {
	return r.data
}

func (r *Reader) Pos() int {
	return r.pos
}

func (r *Reader) ReadLen() int {
	l, off := binary.Varint(r.data[r.pos:])
	r.pos += off
	return int(l)
}

func (r *Reader) Copy(ptr unsafe.Pointer, length int) {
	memmove(ptr, unsafe.Pointer(&r.data[r.pos]), uintptr(length))
	r.pos += length
}

func (r *Reader) Read(length int) uintptr {
	ptr := uintptr(unsafe.Pointer(&r.data[r.pos]))
	r.pos += length
	return ptr
}

type Writer struct {
	data []byte
	pos  int
}

func NewWriter() *Writer {
	return &Writer{data: make([]byte, 16)}
}

func (w *Writer) grow(length int) {
	rem := len(w.data) - w.pos
	if rem >= length {
		w.data = w.data[:len(w.data)]
		return
	}

	w.data = append(w.data, make([]byte, cap(w.data))...)
}

func (w *Writer) Data() []byte {
	return w.data
}

func (w *Writer) WriteLen(length int) {
	w.grow(9)
	w.pos += binary.PutVarint(w.data[w.pos:], int64(length))
}

func (w *Writer) Copy(ptr unsafe.Pointer, length int) {
	w.grow(length)
	memmove(unsafe.Pointer(&w.data[w.pos]), ptr, uintptr(length))
	w.pos += length
}
