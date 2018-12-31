package bstruct

import (
	"bytes"
	"encoding/binary"
	"testing"
)

type small struct {
	A       uint32 `endian:"little"`
	Test1   [4]byte
	B, C, D int16
	Length  int32 `endian:"big"`
	Test2   [4]byte
}

type big struct {
	A       uint32 `endian:"little"`
	Test1   [512]byte
	B, C, D int16
	Length  int32 `endian:"big"`
	Test2   [4]byte
}

type basicCover struct {
	Byte   byte
	Uint16 uint16
	Uint32 uint32
	Uint64 uint64
	Int8   int8
	Int16  int16 `endian:"big"`
	Int32  int32

	K struct {
		Int64 int64 `skip:"r"`
	}
	Array1 [5]byte `rdn:"root.Byte = 4; read(4); if (true) root.Uint16 = 8; if (root.Byte) root.Uint32 = 16;"`

	An struct {
		Drray [4]byte
		Array []byte   `length:"current.Drray[0]/3"`
		Brray []uint16 `size:"4+2"`
	}

	M []struct {
		Int64 int64
	} `length:"4"`

	Float32 float32
	Float64 float64
	Array2  []byte
}

var testbytes = []byte{
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
	1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0, 1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 1, 2, 3, 4, 1, 2, 3, 4, 0, 0,
}

var smallbench = small{}
var bigbench = big{}

var st = MustNew(smallbench)
var bt = MustNew(bigbench)
var dec = NewDecoder()
var enc = NewEncoder()

func BenchmarkSmallDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dec.Rd = bytes.NewReader(testbytes)
		if e := dec.Decode(st, &smallbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkStdSmallDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var rd = bytes.NewReader(testbytes)
		if e := binary.Read(rd, binary.BigEndian, &smallbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkBigDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dec.Rd = bytes.NewReader(testbytes)
		if e := dec.Decode(bt, &bigbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkStdBigDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var rd = bytes.NewReader(testbytes)
		if e := binary.Read(rd, binary.BigEndian, &bigbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkSmallEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		enc.Wt = &bytes.Buffer{}
		if e := enc.Encode(st, &smallbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkStdSmallEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if e := binary.Write(&buf, binary.BigEndian, smallbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkBigEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		enc.Wt = &bytes.Buffer{}
		if e := enc.Encode(bt, &bigbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkStdBigEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if e := binary.Write(&buf, binary.BigEndian, bigbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func TestApi(t *testing.T) {
	var cover = &basicCover{}
	var ct = MustNew(cover)

	dec.Rd = bytes.NewReader(testbytes)
	if e := dec.Decode(ct, cover); e != nil {
		t.Fatalf("%+v\n", e)
	}

	enc.Wt = &bytes.Buffer{}
	if e := enc.Encode(ct, cover); e != nil {
		t.Fatalf("%+v\n", e)
	}

	t.Logf("%+v\n", cover)
	if cover.Byte != 4 || cover.Uint16 != 8 || cover.Uint32 != 16 {
		t.Fatal("set variable is not working")
	}

	if len(cover.An.Array) != int(cover.An.Drray[0])/3 || len(cover.M) != 4 {
		t.Fatal("length program is not working")
	}

	if len(cover.An.Brray) != 3 {
		t.Fatal("size program is not working")
	}
}
