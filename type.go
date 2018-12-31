package bstruct

import vm "github.com/xhebox/bstruct/tinyvm"

type Type struct {
	kind Kind

	slice_k     bool
	slice_mode  sliceMode
	slice_elem  *Type
	slice_extra *vm.Prog

	struct_num  int
	struct_elem []*Field
}

// variable type
func (this *Type) Kind() Kind {
	return this.kind
}

// if this's a slice/array, will return its child Type
// or nil
func (this *Type) Elem() *Type {
	return this.slice_elem
}

// num of field
func (this *Type) NumField() int {
	return this.struct_num
}

// for struct
// return its child Type by index
// or nil
func (this *Type) FieldByIndex(i int) *Field {
	return this.struct_elem[i]
}

// for struct
// return a field by name
// or nil
func (this *Type) FieldByName(s string) *Field {
	for _, v := range this.struct_elem {
		if v.name == s {
			return v
		}
	}

	return nil
}
