package bstruct

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/xhebox/bstruct/byteorder"
)

// just like New(), new type or panic
func MustNew(data interface{}) *Type {
	t, e := New(data)
	if e != nil {
		panic(e)
	}

	return t
}

// generate a New Type based on what you have passed. *Type is used by Decoder and Encoder to write/read.

// pure type is not allowed to be passed as an argument, but there's a trick to get Type from pure type. that is to pass a pointer of nil, e.g. (*Type)(nil), so you do not need to have a instance.
//
// script is supported by a custom c-like language. context is kept from Type's construction to its destruction.
//
// and you can, of course, generate Type for non-struct variable.
//
// most importantly, always generate Type by this function instead of new(Type) or Type{}
func New(data interface{}) (*Type, error) {
	{
		a := 0xABCD
		if uint8(a) == 0xAB {
			HostEndian = byteorder.BigEndian
		} else {
			HostEndian = byteorder.LittleEndian
		}
	}

	v := reflect.ValueOf(data).Type()
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Array, reflect.Slice, reflect.String, reflect.Struct:
		return genType(v)
	default:
		return nil, errors.Errorf("unsupported type %v", v.Kind())
	}
}

func genType(typ reflect.Type) (r *Type, e error) {
	switch kind := typ.Kind(); kind {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.String:
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
			errors.WithStack(e)
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
			kind:       Kind(kind),
			struct_num: typ.NumField(),
		}

		for i, j := 0, typ.NumField(); i < j; i++ {
			var t *Field

			subfield := typ.Field(i)

			// unexported
			if len(subfield.PkgPath) != 0 {
				t = &Field{rtype: &Type{kind: Invalid}}
			} else {
				t, e = genField(subfield, typ)
				if e != nil {
					e = errors.WithStack(e)
					return
				}
			}

			t.name = subfield.Name
			r.struct_elem = append(r.struct_elem, t)
		}
	default:
		e = errors.Errorf("unsupported type %s", kind)
		return
	}

	return r, nil
}

func genField(field reflect.StructField, prt reflect.Type) (r *Field, e error) {
	r, e = newField(field)

	switch kind := field.Type.Kind(); kind {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		sz := basicsize(Kind(kind))
		if r.align < sz {
			r.align = sz
		}

		r.rtype = &Type{kind: Kind(kind)}
	case reflect.Interface, reflect.Uint, reflect.Int:
		r.rtype = &Type{kind: Invalid}

		if len(r.prog["type"]) == 0 {
			e = errors.New("interface{}/int/uint field must have a type program")
			return
		}
	case reflect.String:
		r.rtype = &Type{kind: String}
	case reflect.Slice, reflect.Array:
		k := field.Type.Elem()
		kkind := k.Kind()

		sz := basicsize(Kind(kkind))
		if r.align < sz {
			r.align = sz
		}

		var elem *Type
		elem, e = genType(k)
		if e != nil {
			errors.WithStack(e)
			return
		}

		r.rtype = &Type{
			kind:       Kind(kind),
			slice_mode: SliceModeEOF,
			slice_elem: elem,
		}

		if kind == reflect.Array {
			r.rtype.slice_mode = SliceModeLen
		}

		if length := field.Tag.Get("length"); len(length) != 0 {
			r.rtype.slice_mode = SliceModeLen
			r.rtype.slice_extra = length
		} else if size := field.Tag.Get("size"); len(size) != 0 {
			r.rtype.slice_mode = SliceModeSize
			r.rtype.slice_extra = size
		}
	case reflect.Struct:
		r.rtype = &Type{
			kind:       Kind(kind),
			struct_num: field.Type.NumField(),
		}

		for i, j := 0, field.Type.NumField(); i < j; i++ {
			var t *Field

			sfield := field.Type.Field(i)

			// unexported
			if len(sfield.PkgPath) != 0 || (r.flag&FlagSkip == FlagSkipr|FlagSkipw) {
				t = &Field{rtype: &Type{kind: Invalid}}
			} else {
				t, e = genField(sfield, field.Type)
				if e != nil {
					e = errors.WithStack(e)
					return
				}
			}

			t.name = sfield.Name
			r.rtype.struct_elem = append(r.rtype.struct_elem, t)
		}
	default:
		e = errors.New("unsupported type")
	}

	return
}
