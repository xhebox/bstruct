package bstruct

import (
	"bytes"
	"encoding/binary"
	"testing"
)

type small struct {
	A       uint32  `endian:"big"`
	Test1   [4]byte `endian:"big"`
	B, C, D int16   `endian:"big"`
	Length  int32   `endian:"big"`
	Test2   [4]byte `endian:"big"`
}

type big struct {
	A       uint32     `endian:"big"`
	Test1   [128]int32 `endian:"big"`
	B, C, D int16      `endian:"big"`
	Length  int32      `endian:"big"`
	Test2   [4]byte    `endian:"big"`
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
	Array1 [5]byte `rdn:"Array1_test"`

	An struct {
		Drray [4]byte
		Array []byte   `length:"Array_length"`
		Brray []uint16 `size:"An_size"`
	}

	M []struct {
		Int64 int64
	} `length:"M_length"`

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

func BenchmarkSmallDecode(b *testing.B) {
	var st = MustNew(smallbench)
	var dec = NewDecoder()
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
	var bt = MustNew(bigbench)
	var dec = NewDecoder()
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
	var st = MustNew(smallbench)
	buf := &bytes.Buffer{}
	var enc = NewEncoder()
	enc.Wt = buf
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if e := enc.Encode(st, &smallbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkStdSmallEncode(b *testing.B) {
	buf := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if e := binary.Write(buf, binary.BigEndian, smallbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkBigEncode(b *testing.B) {
	var bt = MustNew(bigbench)
	var enc = NewEncoder()
	buf := &bytes.Buffer{}
	enc.Wt = buf
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if e := enc.Encode(bt, &bigbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkStdBigEncode(b *testing.B) {
	buf := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if e := binary.Write(buf, binary.BigEndian, bigbench); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkApiDecode(b *testing.B) {
	var cover = &basicCover{}
	var ct = MustNew(cover)
	var dec = NewDecoder()

	dec.Runner.Register("An_size", func(...interface{}) interface{} {
		return 4 + 2
	})

	dec.Runner.Register("Array_length", func(s ...interface{}) interface{} {
		r := s[0].(*basicCover)

		return int(r.An.Drray[0] / 3)
	})

	dec.Runner.Register("M_length", func(...interface{}) interface{} {
		return 4
	})

	dec.Runner.Register("Array1_test", func(s ...interface{}) interface{} {
		r := s[1].(basicCover)

		r.Byte = 4
		r.Uint16 = 8
		r.Uint32 = 16
		return nil
	})

	for i := 0; i < b.N; i++ {
		dec.Rd = bytes.NewReader(testbytes)
		if e := dec.Decode(ct, cover); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func BenchmarkApiEncode(b *testing.B) {
	var cover = &basicCover{}
	var ct = MustNew(cover)
	var enc = NewEncoder()

	for i := 0; i < b.N; i++ {
		enc.Wt = &bytes.Buffer{}
		if e := enc.Encode(ct, cover); e != nil {
			b.Fatalf("%+v\n", e)
		}
	}
}

func TestApi(t *testing.T) {
	var cover = &basicCover{}
	var ct = MustNew(cover)
	var dec = NewDecoder()
	var enc = NewEncoder()

	dec.Runner.Register("An_size", func(...interface{}) interface{} {
		return 4 + 2
	})

	dec.Runner.Register("Array_length", func(s ...interface{}) interface{} {
		r := s[0].(*basicCover)

		return int(r.An.Drray[0] / 3)
	})

	dec.Runner.Register("M_length", func(...interface{}) interface{} {
		return 4
	})

	dec.Runner.Register("Array1_test", func(s ...interface{}) interface{} {
		r := s[0].(*basicCover)

		r.Byte = 4
		r.Uint16 = 8
		r.Uint32 = 16
		return nil
	})

	dec.Rd = bytes.NewReader(testbytes)
	if e := dec.Decode(ct, cover); e != nil {
		t.Fatalf("%+v\n", e)
	}

	enc.Wt = &bytes.Buffer{}
	enc.Runner.Copy(dec.Runner)
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
