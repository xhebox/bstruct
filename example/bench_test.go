package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xhebox/bstruct"
)

func makeData(b *testing.B) []byte {
	f := &Struct1{}
	wt := bstruct.NewWriter()
	f.Encode(wt)
	return wt.Data()
}

func TestAff(t *testing.T) {
	f := &Struct1{
		B: nil,
		F: true,
		D: "gg",
		G: []string{
			"1",
			"3",
			"154",
		},
		E: Slice1{
			{E: "1"},
		},
	}
	g := &Struct1{}
	t.Logf("%+v\n%+v\n", f, g)

	wt := bstruct.NewWriter()
	f.Encode(wt)
	rd := bstruct.NewReader(wt.Data())
	g.Decode(rd)
	require.Equal(t, f, g)
	t.Logf("%+v\n%+v\n", f, g)
}

func BenchmarkEncode(b *testing.B) {
	f := &Struct1{
		G: []string{
			"1",
			"3",
			"154",
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wt := bstruct.NewWriter()
		f.Encode(wt)
	}
}

func BenchmarkMarshal(b *testing.B) {
	f := &Struct1{
		G: []string{
			"1",
			"3",
			"154",
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(f)
		require.NoError(b, err)
	}
}

func BenchmarkDecode(b *testing.B) {
	f := &Struct1{
		G: []string{
			"1",
			"3",
			"154",
		},
	}
	wt := bstruct.NewWriter()
	f.Encode(wt)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rd := bstruct.NewReader(wt.Data())
		f.Decode(rd)
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	f := &Struct1{
		G: []string{
			"1",
			"3",
			"154",
		},
	}
	data, err := json.Marshal(f)
	require.NoError(b, err)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = json.Unmarshal(data, f)
		require.NoError(b, err)
	}
}
