package main

//go:generate go run . -o gen.go -p main
//go:generate go fmt .

import (
	"bytes"
	"flag"
	"log"
	"os"

	. "github.com/xhebox/bstruct"
)

func main() {
	out := flag.String("o", "gen.go", "output file")
	pak := flag.String("p", "bench", "package name")
	flag.Parse()

	buf := new(bytes.Buffer)
	enc := NewBuilder().Getter(true).Setter(true)
	New(FieldStruct).
		Reg(enc, "Struct1").
		Comment("Struct1 is fff").
		Add("A", "gfgf", false, New(FieldBool)).
		Add("B", "", false, NewSlice(New(FieldBool))).
		Add("C", "", false,
			New(FieldStruct).
				Reg(enc, "Struct2").
				Add("A", "", false, New(FieldBool)),
		).
		Add("D", "", false, NewString()).
		Add("G", "", false, NewSlice(NewString())).
		Add("E", "", false,
			NewSlice(New(FieldStruct).
				Reg(enc, "Slice1").
				Add("E", "", false, NewString())),
		).
		Add("fieldGerrrccontrol", "", true, New(FieldBool))
	enc.Process()
	enc.Print(buf, *pak)
	if err := os.WriteFile(*out, buf.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}
