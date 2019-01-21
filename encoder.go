package bstruct

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/xhebox/bstruct/byteorder"
	vm "github.com/xhebox/bstruct/tinyvm"
)

func str_bytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

// Encoder has three exported field.
//
// 1. Wt: every time you want to write something new, you need to refresh the writer.
//
// 2. Endian: when you want to change the default Endian on the fly
//
// 3. VM: when you want to pass an external variable, it will be reflect-based whatever type it is
type Encoder struct {
	Wt     io.Writer
	Endian byteorder.ByteOrder
	VM     *vm.VM
	align  int
	buf    []byte
}

// just like New() *Type, always create encoder by this function
func NewEncoder() *Encoder {
	{
		a := 0xABCD
		if uint8(a) == 0xAB {
			HostEndian = byteorder.BigEndian
		} else {
			HostEndian = byteorder.LittleEndian
		}
	}

	enc := &Encoder{
		Endian: HostEndian,
		VM: &vm.VM{
			Endian: HostEndian,
		},
		buf: make([]byte, 16),
	}
	enc.VM.Init(256, 256)
	enc.VM.Set("view", func(x ...interface{}) {
		for _, v := range x {
			fmt.Println(v)
		}
	})
	enc.VM.Set("fill", func(x int64) {
		buf := make([]byte, x)
		if _, e := enc.Wt.Write(buf); e != nil {
			panic(e)
		}
	})

	return enc
}

// pass the generated *Type, and a pointer to data
func (t *Encoder) Encode(w *Type, data interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(data))
	t.VM.Root = v
	t.align = basicsize(w.kind)
	return t.encode(w, v)
}

func (t *Encoder) encode(w *Type, v reflect.Value) error {
	switch w.kind {
	case Invalid:
	case String:
		if _, e := t.Wt.Write(str_bytes(v.String())); e != nil {
			return errors.Wrapf(e, "can not write string")
		}
	case Varint:
		n := t.Endian.PutUVarint(t.buf, v.Uint())
		if _, e := t.Wt.Write(t.buf[:n]); e != nil {
			return errors.Wrapf(e, "can not write uvarint")
		}
	case UVarint:
		n := t.Endian.PutVarint(t.buf, v.Int())
		if _, e := t.Wt.Write(t.buf[:n]); e != nil {
			return errors.Wrapf(e, "can not write varint")
		}
	case Bool:
		t.Endian.PutBool(t.buf, v.Bool())

		if t.align > 1 {
			for k, e := 1, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write bool")
		}
	case Int8:
		t.Endian.PutInt8(t.buf, int8(v.Int()))

		if t.align > 1 {
			for k, e := 1, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write int8")
		}
	case Int16:
		t.Endian.PutInt16(t.buf, int16(v.Int()))

		if t.align > 2 {
			for k, e := 2, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write int16")
		}
	case Int32:
		t.Endian.PutInt32(t.buf, int32(v.Int()))

		if t.align > 4 {
			for k, e := 4, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write int32")
		}
	case Int64:
		t.Endian.PutInt64(t.buf, v.Int())

		if t.align > 8 {
			for k, e := 8, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write int64")
		}
	case Uint8:
		t.Endian.PutUint8(t.buf, uint8(v.Uint()))

		if t.align > 1 {
			for k, e := 1, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write uint8")
		}
	case Uint16:
		t.Endian.PutUint16(t.buf, uint16(v.Uint()))

		if t.align > 2 {
			for k, e := 2, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write uint16")
		}
	case Uint32:
		t.Endian.PutUint32(t.buf, uint32(v.Uint()))

		if t.align > 4 {
			for k, e := 4, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write uint32")
		}
	case Uint64:
		t.Endian.PutUint64(t.buf, v.Uint())

		if t.align > 8 {
			for k, e := 8, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write uint64")
		}
	case Float32:
		t.Endian.PutFloat32(t.buf, float32(v.Float()))

		if t.align > 4 {
			for k, e := 4, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write float32")
		}
	case Float64:
		t.Endian.PutFloat64(t.buf, v.Float())

		if t.align > 8 {
			for k, e := 8, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write float64")
		}
	case Array, Slice:
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

			switch kind {
			case Bool:
				slice := make([]bool, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for k := 0; k < l; k++ {
					if slice[k] {
						t.buf[k] = 1
					}
				}
			case Int8:
				slice := make([]int8, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for k := 0; k < l; k++ {
					t.buf[k] = byte(slice[k])
				}
			case Uint8:
				reflect.Copy(reflect.ValueOf(t.buf), v)
			case Uint16:
				slice := make([]uint16, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for k := 0; k < l; k++ {
					t.Endian.PutUint16(t.buf[k*sz:], slice[k])
				}
			case Uint32:
				slice := make([]uint32, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for k := 0; k < l; k++ {
					t.Endian.PutUint32(t.buf[k*sz:], slice[k])
				}
			case Uint64:
				slice := make([]uint64, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for k := 0; k < l; k++ {
					t.Endian.PutUint64(t.buf[k*sz:], slice[k])
				}
			case Int16:
				slice := make([]int16, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for k := 0; k < l; k++ {
					t.Endian.PutUint16(t.buf[k*sz:], uint16(slice[k]))
				}
			case Int32:
				slice := make([]int32, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for k := 0; k < l; k++ {
					t.Endian.PutUint32(t.buf[k*sz:], uint32(slice[k]))
				}
			case Int64:
				slice := make([]int64, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for k := 0; k < l; k++ {
					t.Endian.PutUint64(t.buf[k*sz:], uint64(slice[k]))
				}
			case Float32:
				slice := make([]float32, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for k := 0; k < l; k++ {
					t.Endian.PutFloat32(t.buf[k*sz:], slice[k])
				}
			case Float64:
				slice := make([]float64, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for k := 0; k < l; k++ {
					t.Endian.PutFloat64(t.buf[k*sz:], slice[k])
				}
			}

			if _, e := t.Wt.Write(t.buf[:m]); e != nil {
				return errors.WithStack(e)
			}
		} else {
			for cnt := 0; cnt < l; cnt++ {
				t.VM.K = int64(cnt)
				if e := t.encode(elem, v.Index(cnt)); e != nil {
					return errors.Wrapf(e, "can not execute encode for elem[%d]", cnt)
				}
			}
		}
	case Struct:
		oricur := t.VM.Current

		t.VM.Current = v
		for k, f := range w.struct_elem {
			var fw = f.rtype
			var fv = v.Field(k)

			if f.tpm != nil {
				if e := t.VM.Exec(f.tpm); e != nil {
					return errors.Wrapf(e, "can not execute field tpm program")
				}

				rtype, ok := Types[t.VM.Ret().ToString()]
				if !ok {
					return errors.New("can not resolve type casting")
				}
				f.rtype = rtype
			}

			if f.wtm != nil {
				if e := t.VM.Exec(f.wtm); e != nil {
					return errors.Wrapf(e, "can not execute field pre program")
				}
			}

			if len(f.name) != 0 {
				if f.flag&FlagSkipw == 0 {
					var oriend = t.Endian

					if f.flag&FlagCusEnd != 0 {
						if f.flag&FlagBig != 0 {
							t.Endian = byteorder.BigEndian
						} else {
							t.Endian = byteorder.LittleEndian
						}
					}

					t.align = f.align

					if e := t.encode(fw, fv); e != nil {
						return errors.Wrapf(e, "can not execute encode for field [%s]", f.Name())
					}

					t.Endian = oriend
				}
			}

			if f.wtn != nil {
				if e := t.VM.Exec(f.wtn); e != nil {
					return errors.Wrapf(e, "can not execute field post program")
				}
			}
		}

		t.VM.Current = oricur
	default:
		return errors.Errorf("unsupported type: %v\n", w.kind)
	}

	return nil
}
