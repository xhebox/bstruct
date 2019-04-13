package bstruct

import (
	"io"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/xhebox/bstruct/byteorder"
)

const (
	max   = int(^uint(0) >> 1)
	max64 = int64(^uint64(0) >> 1)
)

type SliceMode uint8

const (
	SliceModeLen SliceMode = iota
	SliceModeEOF
	SliceModeSize
)

var (
	// 4 is the bound where reflect way is still faster than my speed trick for large array/slice
	// but that is the optimized value on my machine, you could bench your own value
	SliceAccelThreshold = 4

	// the initial guess length when it's modeEOF
	SliceInitLen = 256
)

func basicsize(k Kind) int {
	switch k {
	case Bool, Int8, Uint8:
		return 1
	case Int16, Uint16:
		return 2
	case Int32, Uint32, Float32:
		return 4
	case Int64, Uint64, Float64:
		return 8
	default:
		return 0
	}
}

// Decoder has three exported field.
//
// 1. Rd: every time you want to read something new, you need to refresh the reader.
//
// 2. Endian: when you want to change the default Endian on the fly
type Decoder struct {
	Rd     io.Reader
	Endian byteorder.ByteOrder
	buf    []byte
	root   interface{}
	*runner
}

// just like New() *Type, always create decoder by this function
func NewDecoder() *Decoder {
	{
		a := 0xABCD
		if uint8(a) == 0xAB {
			HostEndian = byteorder.BigEndian
		} else {
			HostEndian = byteorder.LittleEndian
		}
	}

	dec := &Decoder{
		Endian: HostEndian,
		runner: &runner{
			progs: map[string]func(...interface{}) interface{}{},
		},
		buf: make([]byte, 16),
	}

	return dec
}

// pass the generated *Type, and a pointer to data
func (t *Decoder) Decode(w *Type, data interface{}) error {
	t.root = data
	return t.decode(w, basicsize(w.kind), reflect.Indirect(reflect.ValueOf(data)))
}

func (t *Decoder) decode(w *Type, align int, v reflect.Value) error {
	switch w.kind {
	case Invalid:
	case String:
	case UVarint:
		n, e := t.Endian.UVarint(t.Rd)
		if e != nil {
			return errors.Wrapf(e, "can not read uvarint")
		}

		v.SetUint(n)
	case Varint:
		n, e := t.Endian.Varint(t.Rd)
		if e != nil {
			return errors.Wrapf(e, "can not read varint")
		}

		v.SetInt(n)
	case Bool:
		if _, e := t.Rd.Read(t.buf[:align]); e != nil {
			return errors.Wrapf(e, "can not read bool")
		}

		v.SetBool(t.Endian.Bool(t.buf))
	case Int8:
		if _, e := t.Rd.Read(t.buf[:align]); e != nil {
			return errors.Wrapf(e, "can not read int8")
		}

		v.SetInt(int64(t.Endian.Int8(t.buf)))
	case Int16:
		if _, e := t.Rd.Read(t.buf[:align]); e != nil {
			return errors.Wrapf(e, "can not read int16")
		}

		v.SetInt(int64(t.Endian.Int16(t.buf)))
	case Int32:
		if _, e := t.Rd.Read(t.buf[:align]); e != nil {
			return errors.Wrapf(e, "can not read int32")
		}

		v.SetInt(int64(t.Endian.Int32(t.buf)))
	case Int64:
		if _, e := t.Rd.Read(t.buf[:align]); e != nil {
			return errors.Wrapf(e, "can not read int64")
		}

		v.SetInt(t.Endian.Int64(t.buf))
	case Uint8:
		if _, e := t.Rd.Read(t.buf[:align]); e != nil {
			return errors.Wrapf(e, "can not read uint8")
		}

		v.SetUint(uint64(t.Endian.Uint8(t.buf)))
	case Uint16:
		if _, e := t.Rd.Read(t.buf[:align]); e != nil {
			return errors.Wrapf(e, "can not read uint16")
		}

		v.SetUint(uint64(t.Endian.Uint16(t.buf)))
	case Uint32:
		if _, e := t.Rd.Read(t.buf[:align]); e != nil {
			return errors.Wrapf(e, "can not read uint32")
		}

		v.SetUint(uint64(t.Endian.Uint32(t.buf)))
	case Uint64:
		if _, e := t.Rd.Read(t.buf[:align]); e != nil {
			return errors.Wrapf(e, "can not read uint64")
		}

		v.SetUint(t.Endian.Uint64(t.buf))
	case Float32:
		if _, e := t.Rd.Read(t.buf[:align]); e != nil {
			return errors.Wrapf(e, "can not read float32")
		}

		v.SetFloat(float64(t.Endian.Float32(t.buf)))
	case Float64:
		if _, e := t.Rd.Read(t.buf[:align]); e != nil {
			return errors.Wrapf(e, "can not read float64")
		}

		v.SetFloat(t.Endian.Float64(t.buf))
	case Array, Slice:
		var ord = t.Rd
		var mode = w.slice_mode

		if len(w.slice_extra) != 0 {
			switch w.slice_mode {
			case SliceModeLen:
				l, ok := t.runner.exec(w.slice_extra, t.root).(int)
				if !ok {
					return errors.Errorf("can not execute length program")
				}

				if l > 0 {
					v.Set(reflect.MakeSlice(v.Type(), l, l))
				} else if l == 0 {
					return nil
				} else {
					return errors.Errorf("length program returned a negative %d", l)
				}
			case SliceModeSize:
				l, ok := t.runner.exec(w.slice_extra, t.root).(int)
				if !ok {
					return errors.Errorf("can not execute size program")
				}

				if l > 0 {
					sz := basicsize(w.slice_elem.Kind())
					if sz <= 0 {
						t.Rd = io.LimitReader(ord, int64(l))
						mode = SliceModeEOF

						v.Set(reflect.MakeSlice(v.Type(), l, l))
					} else {
						cnt := int(l) / sz
						v.Set(reflect.MakeSlice(v.Type(), cnt, cnt))

						mode = SliceModeLen
					}
				} else if l == 0 {
					return nil
				} else {
					mode = SliceModeEOF
				}
			}
		}

		if mode == SliceModeEOF && v.Len() != 0 {
			mode = SliceModeLen
		}

		switch mode {
		case SliceModeLen:
			l := v.Len()
			elem := w.slice_elem
			kind := elem.kind

			if kind.IsBasic() && l > SliceAccelThreshold {
				sz := basicsize(kind)
				m := l * sz

				if n := cap(t.buf); n < m {
					for n < m {
						n *= 2
					}
					t.buf = make([]byte, n)
				}

				if _, e := t.Rd.Read(t.buf[:m]); e != nil {
					return errors.WithStack(e)
				}

				switch kind {
				case Bool:
					slice := make([]bool, l)

					for k := 0; k < l; k++ {
						slice[k] = t.Endian.Bool(t.buf[k:])
					}

					reflect.Copy(v, reflect.ValueOf(slice))
				case Int8:
					slice := make([]int8, l)

					for k := 0; k < l; k++ {
						slice[k] = t.Endian.Int8(t.buf[k:])
					}

					reflect.Copy(v, reflect.ValueOf(slice))
				case Int16:
					slice := make([]int16, l)

					for k := 0; k < l; k++ {
						slice[k] = t.Endian.Int16(t.buf[k*sz:])
					}

					reflect.Copy(v, reflect.ValueOf(slice))
				case Int32:
					slice := make([]int32, l)

					for k := 0; k < l; k++ {
						slice[k] = t.Endian.Int32(t.buf[k*sz:])
					}

					reflect.Copy(v, reflect.ValueOf(slice))
				case Int64:
					slice := make([]int64, l)

					for k := 0; k < l; k++ {
						slice[k] = t.Endian.Int64(t.buf[k*sz:])
					}

					reflect.Copy(v, reflect.ValueOf(slice))
				case Uint8:
					reflect.Copy(v, reflect.ValueOf(t.buf))
				case Uint16:
					slice := make([]uint16, l)

					for k := 0; k < l; k++ {
						slice[k] = t.Endian.Uint16(t.buf[k*sz:])
					}

					reflect.Copy(v, reflect.ValueOf(slice))
				case Uint32:
					slice := make([]uint32, l)

					for k := 0; k < l; k++ {
						slice[k] = t.Endian.Uint32(t.buf[k*sz:])
					}

					reflect.Copy(v, reflect.ValueOf(slice))
				case Uint64:
					slice := make([]uint64, l)

					for k := 0; k < l; k++ {
						slice[k] = t.Endian.Uint64(t.buf[k*sz:])
					}

					reflect.Copy(v, reflect.ValueOf(slice))
				case Float32:
					slice := make([]float32, l)

					for k := 0; k < l; k++ {
						slice[k] = t.Endian.Float32(t.buf[k*sz:])
					}

					reflect.Copy(v, reflect.ValueOf(slice))
				case Float64:
					slice := make([]float64, l)

					for k := 0; k < l; k++ {
						slice[k] = t.Endian.Float64(t.buf[k*sz:])
					}

					reflect.Copy(v, reflect.ValueOf(slice))
				}
			} else {
				for cnt := 0; cnt < l; cnt++ {
					if e := t.decode(elem, align, v.Index(cnt)); e != nil {
						return errors.Wrapf(e, "can not execute decode for elem[%d]", cnt)
					}
				}
			}
		case SliceModeEOF:
			vtype := v.Type()
			elem := w.slice_elem
			v.Set(reflect.MakeSlice(vtype, SliceInitLen, SliceInitLen))

			cnt := 0
			for nm := v.Len(); cnt < max; {
				for ; cnt < nm; cnt++ {
					if e := t.decode(elem, align, v.Index(cnt)); e != nil {
						if strings.HasSuffix(e.Error(), "EOF") {
							v.SetLen(cnt)
							t.Rd = ord
							return nil
						}

						return errors.Wrapf(e, "can not decode elem[%d]", cnt)
					}
				}

				var nv = reflect.MakeSlice(vtype, v.Len()*2, v.Len()*2)
				reflect.Copy(nv, v)
				v.Set(nv)
				nm = v.Len()
			}
			v.SetLen(max)
		default:
			panic("internal error")
		}

		t.Rd = ord
	case Struct:
		for k, f := range w.struct_elem {
			var fw = f.rtype
			var fv = v.Field(k)

			if l := len(f.prog["type"]); l != 0 {
				typ := ""

				if f.prog["type"][0] == '\'' && f.prog["type"][l-1] == '\'' {
					typ = f.prog["type"][1:l]
				} else {
					var ok bool
					typ, ok = t.runner.exec(f.prog["type"], t.root).(string)
					if !ok {
						return errors.Errorf("can not execute type program")
					}
				}

				rtype, ok := Types[typ]
				if !ok {
					return errors.New("can not resolve type casting")
				}

				fw = rtype
				f.align = basicsize(rtype.kind)
			}

			if len(f.prog["rdm"]) != 0 {
				e, ok := t.runner.exec(f.prog["rdm"], fv.Interface(), t.root).(error)
				if ok {
					return errors.Errorf("can not execute rdm program: %+v", e)
				}
			}

			if len(f.name) != 0 {
				if f.flag&FlagSkipr == 0 {
					var oriend = t.Endian

					if f.flag&FlagCusEnd != 0 {
						if f.flag&FlagBig != 0 {
							t.Endian = byteorder.BigEndian
						} else {
							t.Endian = byteorder.LittleEndian
						}
					}

					if e := t.decode(fw, f.align, fv); e != nil {
						return errors.Wrapf(e, "can not execute decode for field [%s]", f.Name())
					}

					t.Endian = oriend
				}
			}

			if len(f.prog["rdn"]) != 0 {
				e, ok := t.runner.exec(f.prog["rdn"], fv.Interface(), t.root).(error)
				if ok {
					return errors.Errorf("can not execute rdn program: %+v", e)
				}
			}
		}
	default:
		return errors.Errorf("unsupported type: %v\n", w.kind)
	}

	return nil
}
