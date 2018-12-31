package tinyvm

import (
	"reflect"
	"testing"
)

type A struct {
	Test int
}

var (
	progstr = `3 + 3; "!234" < 5`
	cmp     = NewCompiler()
	prog    = cmp.MustCompile(progstr)
)

var (
	vm   = &VM{}
	root = A{}
)

func TestVM(b *testing.T) {
	root.Test = 3
	vm.Init(64, 128)
	vm.Root = reflect.ValueOf(root)
	vm.Exec(prog)
	b.Logf("%+v\n%v\n", prog.code, vm.Ret().ToString())
}

func BenchmarkVM(b *testing.B) {
	vm.Init(64, 128)
	vm.Root = reflect.ValueOf(root)
	for i := 0; i < b.N; i++ {
		vm.Exec(prog)
	}
}
