package bstruct

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/xhebox/bstruct/byteorder"
)

const (
	MaxAlign = 64
)

var (
	// automatically set when package is inited
	HostEndian = byteorder.HostEndian
)

// just like New(), new type or panic
func MustNew(data interface{}) *Type {
	t, e := New(data)
	if e != nil {
		panic(e)
	}

	return t
}

// generate a New Type based on what you have passed. *Type is used by Decoder and Encoder to read/write.
//
// pure type is not allowed to be passed as an argument, but there's a trick to get Type from pure type. that is to pass a pointer of nil, e.g. (*Type)(nil), so you do not need to have a instance.
//
// !!!PAY ATTENTION!!! always generate Type by this function instead of new(Type) or Type{}
func New(data interface{}) (*Type, error) {
	v := reflect.ValueOf(data).Type()
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return genType(v)
}

func genType(typ reflect.Type) (r *Type, e error) {
	switch kind := typ.Kind(); kind {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		r = &Type{kind: Kind(kind)}
	case reflect.Slice, reflect.Array:
		k := typ.Elem()

		switch k.Kind() {
		case reflect.Slice, reflect.String:
			e = errors.New("slice/string is not allowed to be elem of array/slice, if it's not a field of struct")
			return
		}

		var elem *Type
		elem, e = genType(k)
		if e != nil {
			e = errors.Wrapf(e, "can not generate type")
			return
		}

		r = &Type{
			kind:       Kind(kind),
			slice_mode: SliceModeEOF,
			slice_elem: elem,
		}

		if kind == reflect.Array {
			r.slice_mode = SliceModeLen
		}
	case reflect.Struct:
		r = &Type{
			kind: Kind(kind),
		}

		for i, j := 0, typ.NumField(); i < j; i++ {
			var t *Field

			subfield := typ.Field(i)

			// unexported
			if len(subfield.PkgPath) != 0 {
				t = &Field{rtype: &Type{kind: Invalid}, flag: FlagSkip}
			} else {
				t, e = genField(subfield)
				if e != nil {
					e = errors.Wrapf(e, "can not generate field")
					return
				}
			}

			t.name = subfield.Name
			r.struct_elem = append(r.struct_elem, t)
		}
	default:
		e = errors.Wrapf(e, "unsupported type %s", kind)
		return
	}

	return r, nil
}

func genField(field reflect.StructField) (r *Field, e error) {
	r, e = newField(field)
	if e != nil {
		e = errors.Errorf("can not generate field: %v", e)
	}

	switch field.Type.Kind() {
	case reflect.Interface, reflect.Int, reflect.Uint, reflect.String:
		r.rtype = &Type{kind: Invalid}
		return
	}

	r.rtype, e = genType(field.Type)
	if e != nil {
		e = errors.Errorf("can not generate type: %v", e)
	}

	switch field.Type.Kind() {
	case reflect.Slice, reflect.Array:
		if length := field.Tag.Get("length"); len(length) != 0 {
			r.rtype.slice_mode = SliceModeLen
			r.rtype.slice_extra = length
		} else if size := field.Tag.Get("size"); len(size) != 0 {
			r.rtype.slice_mode = SliceModeSize
			r.rtype.slice_extra = size
		}
	}

	return
}
