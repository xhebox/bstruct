package bstruct

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	vm "github.com/xhebox/bstruct/tinyvm"
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
	compiler := vm.NewCompiler()

	{
		a := 0xABCD
		if uint8(a) == 0xAB {
			HostEndian = BigEndian
		} else {
			HostEndian = LittleEndian
		}
		compiler.Endian = HostEndian
	}

	v := reflect.Indirect(reflect.ValueOf(data))
	switch v.Kind() {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Array, reflect.Slice, reflect.String, reflect.Struct:
		compiler.R = v.Type()
		return genType(compiler, compiler.R)
	default:
		return nil, errors.Errorf("unsupported type %v", v.Kind())
	}
}

func genType(compiler *vm.Compiler, cur reflect.Type) (r *Type, e error) {
	switch kind := cur.Kind(); kind {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.String:
		r = &Type{kind: Kind(kind)}
	case reflect.Slice, reflect.Array:
		k := cur.Elem()

		switch k.Kind() {
		case reflect.Slice, reflect.String:
			e = errors.New("slice/string is not allowed to be elem of array/slice")
			return
		}

		var elem *Type
		elem, e = genType(compiler, k)
		if e != nil {
			errors.WithStack(e)
			return
		}

		r = &Type{
			kind:       Kind(kind),
			slice_mode: sliceModeEOF,
			slice_elem: elem,
		}

		if kind == reflect.Array {
			r.slice_mode = sliceModeLen
		}
	case reflect.Struct:
		r = &Type{
			kind:       Kind(kind),
			struct_num: cur.NumField(),
		}

		for i, j := 0, cur.NumField(); i < j; i++ {
			var t *Field

			sfield := cur.Field(i)

			// unexported
			if len(sfield.PkgPath) != 0 {
				t = &Field{rtype: &Type{kind: Invalid}}
			} else {
				t, e = genField(compiler, sfield, cur)
				if e != nil {
					e = errors.WithStack(e)
					return
				}
			}

			t.name = sfield.Name
			r.struct_elem = append(r.struct_elem, t)
		}
	default:
		e = errors.Errorf("unsupported type %s", kind)
		return
	}

	return r, nil
}

func genField(compiler *vm.Compiler, field reflect.StructField, prt reflect.Type) (r *Field, e error) {
	var flag FieldFlag
	var align int
	var tpm, rdm, rdn, wtm, wtn *vm.Prog
	var cur = field.Type
	var tag = field.Tag

	compiler.C = prt

	if len(tag) != 0 {
		if end := tag.Get("endian"); len(end) != 0 {
			switch end {
			case "msb", "big":
				flag |= FlagCusEnd
				flag |= FlagBig
			case "lsb", "little":
				flag |= FlagCusEnd
				flag &^= FlagBig
			}
		}

		{
			skip := tag.Get("skip")
			if strings.Contains(skip, "r") {
				flag |= FlagSkipr
			}
			if strings.Contains(skip, "w") {
				flag |= FlagSkipw
			}
		}

		if alignstr := tag.Get("align"); len(alignstr) != 0 {
			align, e = strconv.Atoi(alignstr)
			if e != nil {
				return
			}

			if align > 16 {
				e = errors.Errorf("align has an upper limit of 16")
				return
			}
		}

		if rdmstr := tag.Get("rdm"); len(rdmstr) != 0 {
			rdm = compiler.MustCompile(rdmstr)
		}

		if rdnstr := tag.Get("rdn"); len(rdnstr) != 0 {
			rdn = compiler.MustCompile(rdnstr)
		}

		if wtmstr := tag.Get("wtm"); len(wtmstr) != 0 {
			wtm = compiler.MustCompile(wtmstr)
		}

		if wtnstr := tag.Get("wtn"); len(wtnstr) != 0 {
			wtn = compiler.MustCompile(wtnstr)
		}

		if tpmstr := tag.Get("type"); len(tpmstr) != 0 {
			tpm = compiler.MustCompile(tpmstr)
		}
	}

	switch kind := cur.Kind(); kind {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		sz := basicsize(Kind(kind))
		if align < sz {
			align = sz
		}

		r = &Field{
			rtype: &Type{kind: Kind(kind)},
			flag:  flag,
			align: align,
			tpm:   tpm,
			rdm:   rdm,
			rdn:   rdn,
			wtm:   wtm,
			wtn:   wtn,
		}
	case reflect.Interface, reflect.Uint, reflect.Int:
		if tpm == nil {
			e = errors.New("interface{}/int/uint field must have a type program")
			return
		}

		r = &Field{
			rtype: &Type{kind: Invalid},
			flag:  flag,
			align: align,
			tpm:   tpm,
			rdm:   rdm,
			rdn:   rdn,
			wtm:   wtm,
			wtn:   wtn,
		}
	case reflect.String:
		r = &Field{
			rtype: &Type{kind: String},
			flag:  flag,
			align: align,
			tpm:   tpm,
			rdm:   rdm,
			rdn:   rdn,
			wtm:   wtm,
			wtn:   wtn,
		}
	case reflect.Slice, reflect.Array:
		k := cur.Elem()
		kkind := k.Kind()

		sz := basicsize(Kind(kkind))
		if align < sz {
			align = sz
		}

		switch kkind {
		case reflect.Slice, reflect.String:
			e = errors.New("slice/string is not allowed to be elem of array/slice")
			return
		}

		var elem *Type
		elem, e = genType(compiler, k)
		if e != nil {
			errors.WithStack(e)
			return
		}
		r = &Field{
			rtype: &Type{
				kind:       Kind(kind),
				slice_mode: sliceModeEOF,
				slice_elem: elem,
			},
			flag:  flag,
			align: align,
			tpm:   tpm,
			rdm:   rdm,
			rdn:   rdn,
			wtm:   wtm,
			wtn:   wtn,
		}

		if kind == reflect.Array {
			r.rtype.slice_mode = sliceModeLen
		}

		if length := tag.Get("length"); len(length) != 0 {
			r.rtype.slice_mode = sliceModeLen
			r.rtype.slice_extra = compiler.MustCompile(length)
		} else if size := tag.Get("size"); len(size) != 0 {
			r.rtype.slice_mode = sliceModeSize
			r.rtype.slice_extra = compiler.MustCompile(size)
		}
	case reflect.Struct:
		r = &Field{
			rtype: &Type{
				kind:       Kind(kind),
				struct_num: cur.NumField(),
			},
			flag:  flag,
			align: align,
			tpm:   tpm,
			rdm:   rdm,
			rdn:   rdn,
			wtm:   wtm,
			wtn:   wtn,
		}

		for i, j := 0, cur.NumField(); i < j; i++ {
			var t *Field

			sfield := cur.Field(i)
			// unexported
			if len(sfield.PkgPath) != 0 ||
				// skiprw, and no prog, it's useless
				((flag&FlagSkip == FlagSkipr|FlagSkipw) &&
					(rdm == nil && rdn == nil && wtm == nil && wtn == nil)) {
				t = &Field{rtype: &Type{kind: Invalid}}
			} else {
				t, e = genField(compiler, sfield, cur)
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
