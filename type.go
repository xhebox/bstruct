package bstruct

import "reflect"

type Type struct {
	kind Kind

	slice_mode  SliceMode
	slice_elem  *Type
	slice_extra string

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
	return len(this.struct_elem)
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

func (this *Type) Size(v reflect.Type) int {
	n := this.Kind().Size()
	if n != 0 {
		return n
	}

	switch this.Kind() {
	case Array:
		return this.slice_elem.Size(v.Elem()) * v.Len()
	case Struct:
		n := 0
		for k, b := range this.struct_elem {
			m := b.Type().Size(v.Field(k).Type)
			if m == 0 {
				return 0
			}
			n += m
		}
		return n
	default:
		return 0
	}
}
