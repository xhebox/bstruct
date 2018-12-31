package bstruct

import vm "github.com/xhebox/bstruct/tinyvm"

// implements field
// normal field implementation
type Field struct {
	rtype *Type
	name  string
	flag  FieldFlag
	align int
	tpm   *vm.Prog
	rdm   *vm.Prog
	rdn   *vm.Prog
	wtm   *vm.Prog
	wtn   *vm.Prog
}

// Field Name
func (this *Field) Name() string {
	return this.name
}

// Field type
func (this *Field) Type() *Type {
	return this.rtype
}

// fieldFlag of the field
func (this *Field) Flag() FieldFlag {
	return this.flag
}

// you're allowed to change fieldFlag on the fly
func (this *Field) SetFlag(flag FieldFlag) {
	this.flag = flag
}

// field align
// field flag only makes sense on basic types or slice of basic types
func (this *Field) Align() int {
	return this.align
}

// align should between [size, 16]
func (this *Field) SetAlign(arg int) {
	sz := basicsize(this.rtype.Kind())
	if arg >= sz && arg <= 16 {
		this.align = arg
	}
}
