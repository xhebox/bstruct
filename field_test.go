package bstruct

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

//func TestMain(m *testing.M) {

func TestMain(t *testing.T) {
	enc := NewBuilder()

	buf := new(strings.Builder)
	struc := New(FieldStruct).
		Reg(enc, "Struct1").
		Add("A", New(FieldBool)).
		Add("B", NewSlice(New(FieldBool))).
		Add("C",
			New(FieldStruct).
				Reg(enc, "Struct2").
				Add("A", New(FieldBool)),
		).
		Add("D", NewString()).
		Add("G", NewSlice(NewString())).
		Add("E",
			NewSlice(New(FieldStruct).
				Add("E", NewString())).
				Reg(enc, "Slice1"),
		).
		Add("F", New(FieldBool))
	require.NotNil(t, struc)
	enc.Process()
	require.NoError(t, enc.Print(buf))
	fmt.Printf(buf.String())
}
