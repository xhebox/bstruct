package bstruct

import (
	"fmt"
	"io"
	"reflect"

	"github.com/pkg/errors"
	"github.com/xhebox/bstruct/byteorder"
	vm "github.com/xhebox/bstruct/tinyvm"
)

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
			HostEndian = BigEndian
		} else {
			HostEndian = LittleEndian
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

func (t *Encoder) bool(b bool) {
	t.Endian.PutBool(t.buf, b)
}

func (t *Encoder) int8(b int8) {
	t.Endian.PutInt8(t.buf, b)
}

func (t *Encoder) int16(b int16) {
	t.Endian.PutInt16(t.buf, b)
}

func (t *Encoder) int32(b int32) {
	t.Endian.PutInt32(t.buf, b)
}

func (t *Encoder) int64(b int64) {
	t.Endian.PutInt64(t.buf, b)
}

func (t *Encoder) uint8(b uint8) {
	t.Endian.PutUint8(t.buf, b)
}

func (t *Encoder) uint16(b uint16) {
	t.Endian.PutUint16(t.buf, b)
}

func (t *Encoder) uint32(b uint32) {
	t.Endian.PutUint32(t.buf, b)
}

func (t *Encoder) uint64(b uint64) {
	t.Endian.PutUint64(t.buf, b)
}

func (t *Encoder) float32(b float32) {
	t.Endian.PutFloat32(t.buf, b)
}

func (t *Encoder) float64(b float64) {
	t.Endian.PutFloat64(t.buf, b)
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
	case Bool:
		t.bool(v.Bool())

		if t.align > 1 {
			for k, e := 1, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write bool")
		}
	case Int8:
		t.int8(int8(v.Int()))

		if t.align > 1 {
			for k, e := 1, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write int8")
		}
	case Int16:
		t.int16(int16(v.Int()))

		if t.align > 2 {
			for k, e := 2, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write int16")
		}
	case Int32:
		t.int32(int32(v.Int()))

		if t.align > 4 {
			for k, e := 4, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write int32")
		}
	case Int64:
		t.int64(v.Int())

		if t.align > 8 {
			for k, e := 8, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write int64")
		}
	case Uint8:
		t.uint8(uint8(v.Uint()))

		if t.align > 1 {
			for k, e := 1, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write uint8")
		}
	case Uint16:
		t.uint16(uint16(v.Uint()))

		if t.align > 2 {
			for k, e := 2, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write uint16")
		}
	case Uint32:
		t.uint32(uint32(v.Uint()))

		if t.align > 4 {
			for k, e := 4, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write uint32")
		}
	case Uint64:
		t.uint64(v.Uint())

		if t.align > 8 {
			for k, e := 8, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write uint64")
		}
	case Float32:
		t.float32(float32(v.Float()))

		if t.align > 4 {
			for k, e := 4, t.align; k < e; k++ {
				t.buf[k] = 0
			}
		}

		if _, e := t.Wt.Write(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not write float32")
		}
	case Float64:
		t.float64(v.Float())

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

			var obuf = t.buf

			switch kind {
			case Bool:
				slice := make([]bool, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for cnt := 0; cnt < l; cnt++ {
					if slice[cnt] {
						obuf[0] = 1
					}
					obuf = obuf[sz:]
				}
			case Int8:
				slice := make([]int8, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for cnt := 0; cnt < l; cnt++ {
					obuf[0] = byte(slice[cnt])
					obuf = obuf[sz:]
				}
			case Uint8:
				reflect.Copy(reflect.ValueOf(t.buf), v)
			case Uint16:
				slice := make([]uint16, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for cnt := 0; cnt < l; cnt++ {
					t.Endian.PutUint16(obuf, slice[cnt])
					obuf = obuf[sz:]
				}
			case Int16:
				slice := make([]int16, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for cnt := 0; cnt < l; cnt++ {
					t.Endian.PutUint16(obuf, uint16(slice[cnt]))
					obuf = obuf[sz:]
				}
			case Uint32:
				slice := make([]uint32, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for cnt := 0; cnt < l; cnt++ {
					t.Endian.PutUint32(obuf, slice[cnt])
					obuf = obuf[sz:]
				}
			case Int32:
				slice := make([]int32, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for cnt := 0; cnt < l; cnt++ {
					t.Endian.PutUint32(obuf, uint32(slice[cnt]))
					obuf = obuf[sz:]
				}
			case Uint64:
				slice := make([]uint64, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for cnt := 0; cnt < l; cnt++ {
					t.Endian.PutUint64(obuf, slice[cnt])
					obuf = obuf[sz:]
				}
			case Int64:
				slice := make([]int64, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for cnt := 0; cnt < l; cnt++ {
					t.Endian.PutUint64(obuf, uint64(slice[cnt]))
					obuf = obuf[sz:]
				}
			case Float32:
				slice := make([]float32, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for cnt := 0; cnt < l; cnt++ {
					t.Endian.PutFloat32(obuf, slice[cnt])
					obuf = obuf[sz:]
				}
			case Float64:
				slice := make([]float64, l)

				reflect.Copy(reflect.ValueOf(slice), v)

				for cnt := 0; cnt < l; cnt++ {
					t.Endian.PutFloat64(obuf, slice[cnt])
					obuf = obuf[sz:]
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
		t.VM.Current = v
		for k, f := range w.struct_elem {
			var fw = f.rtype
			var fv = v.Field(k)

			if f.tpm != nil {
				if e := t.VM.Exec(f.tpm); e != nil {
					return errors.Wrapf(e, "can not execute field tpm program")
				}

				rtype, ok := types[t.VM.Ret().ToString()]
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
							t.Endian = BigEndian
						} else {
							t.Endian = LittleEndian
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
	default:
		return errors.Errorf("unsupported type: %v\n", w.kind)
	}

	return nil
}
