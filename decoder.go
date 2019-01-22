package bstruct

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/xhebox/bstruct/byteorder"
	vm "github.com/xhebox/bstruct/tinyvm"
)

const (
	max   = int(^uint(0) >> 1)
	max64 = int64(^uint64(0) >> 1)
)

type sliceMode uint8

const (
	sliceModeLen sliceMode = iota
	sliceModeEOF
	sliceModeSize
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
//
// 3. VM: when you want to pass an external variable, it will be reflect-based whatever type it is
type Decoder struct {
	Rd     io.Reader
	Endian byteorder.ByteOrder
	VM     *vm.VM
	align  int
	buf    []byte
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
		VM: &vm.VM{
			Endian: HostEndian,
		},
		buf: make([]byte, 16),
	}
	dec.VM.Init(256, 256)
	dec.VM.Set("view", func(x ...interface{}) {
		for _, v := range x {
			fmt.Println(v)
		}
	})
	dec.VM.Set("read", func(x int64) []byte {
		buf := make([]byte, x)
		if _, e := dec.Rd.Read(buf); e != nil {
			panic(e)
		}
		return buf
	})
	dec.VM.Set("discard", func(x int64) {
		if rd, ok := dec.Rd.(*bufio.Reader); ok {
			_, e := rd.Discard(int(x))
			if e != nil {
				panic(e)
			}
		} else {
			_, e := io.CopyN(ioutil.Discard, dec.Rd, x)
			if e != nil {
				panic(e)
			}
		}
	})
	dec.VM.Set("startcount", func() {
		dec.Rd = io.LimitReader(dec.Rd, max64)
	})
	dec.VM.Set("stopcount", func() (r int64) {
		rd := dec.Rd.(*io.LimitedReader)
		r = max64 - rd.N
		dec.Rd = rd.R
		return
	})

	return dec
}

// pass the generated *Type, and a pointer to data
func (t *Decoder) Decode(w *Type, data interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(data))
	t.VM.Root = v
	t.align = basicsize(w.kind)
	return t.decode(w, v)
}

func (t *Decoder) decode(w *Type, v reflect.Value) error {
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
		if _, e := t.Rd.Read(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not read bool")
		}

		v.SetBool(t.Endian.Bool(t.buf))
	case Int8:
		if _, e := t.Rd.Read(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not read int8")
		}

		v.SetInt(int64(t.Endian.Int8(t.buf)))
	case Int16:
		if _, e := t.Rd.Read(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not read int16")
		}

		v.SetInt(int64(t.Endian.Int16(t.buf)))
	case Int32:
		if _, e := t.Rd.Read(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not read int32")
		}

		v.SetInt(int64(t.Endian.Int32(t.buf)))
	case Int64:
		if _, e := t.Rd.Read(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not read int64")
		}

		v.SetInt(t.Endian.Int64(t.buf))
	case Uint8:
		if _, e := t.Rd.Read(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not read uint8")
		}

		v.SetUint(uint64(t.Endian.Uint8(t.buf)))
	case Uint16:
		if _, e := t.Rd.Read(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not read uint16")
		}

		v.SetUint(uint64(t.Endian.Uint16(t.buf)))
	case Uint32:
		if _, e := t.Rd.Read(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not read uint32")
		}

		v.SetUint(uint64(t.Endian.Uint32(t.buf)))
	case Uint64:
		if _, e := t.Rd.Read(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not read uint64")
		}

		v.SetUint(t.Endian.Uint64(t.buf))
	case Float32:
		if _, e := t.Rd.Read(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not read float32")
		}

		v.SetFloat(float64(t.Endian.Float32(t.buf)))
	case Float64:
		if _, e := t.Rd.Read(t.buf[:t.align]); e != nil {
			return errors.Wrapf(e, "can not read float64")
		}

		v.SetFloat(t.Endian.Float64(t.buf))
	case Array, Slice:
		var ord = t.Rd
		var mode = w.slice_mode

		if w.slice_extra != nil {
			switch w.slice_mode {
			case sliceModeLen:
				if e := t.VM.Exec(w.slice_extra); e != nil {
					return errors.Wrapf(e, "can not execute length program")
				}

				l := int(t.VM.Ret().ToInteger())
				if l > 0 {
					v.Set(reflect.MakeSlice(v.Type(), l, l))
				} else if l == 0 {
					return nil
				} else {
					return errors.Errorf("length program returned a negative %d", l)
				}
			case sliceModeSize:
				if e := t.VM.Exec(w.slice_extra); e != nil {
					return errors.Wrapf(e, "can not execute size program")
				}

				l := int(t.VM.Ret().ToInteger())
				if l > 0 {
					sz := basicsize(w.slice_elem.Kind())
					if sz <= 0 {
						t.Rd = io.LimitReader(ord, int64(l))
						mode = sliceModeEOF

						v.Set(reflect.MakeSlice(v.Type(), l, l))
					} else {
						cnt := int(l) / sz
						v.Set(reflect.MakeSlice(v.Type(), cnt, cnt))

						mode = sliceModeLen
					}
				} else if l == 0 {
					return nil
				} else {
					mode = sliceModeEOF
				}
			}
		}

		switch mode {
		case sliceModeLen:
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
					t.VM.K = int64(cnt)
					if e := t.decode(elem, v.Index(cnt)); e != nil {
						return errors.Wrapf(e, "can not execute decode for elem[%d]", cnt)
					}
				}
			}
		case sliceModeEOF:
			vtype := v.Type()
			elem := w.slice_elem
			v.Set(reflect.MakeSlice(vtype, SliceInitLen, SliceInitLen))

			cnt := 0
			for nm := v.Len(); cnt < max; {
				for ; cnt < nm; cnt++ {
					if e := t.decode(elem, v.Index(cnt)); e != nil {
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
				fw = rtype
				f.align = basicsize(rtype.kind)
			}

			if f.rdm != nil {
				if e := t.VM.Exec(f.rdm); e != nil {
					return errors.Wrapf(e, "can not execute field pre program")
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

					t.align = f.align

					if e := t.decode(fw, fv); e != nil {
						return errors.Wrapf(e, "can not execute decode for field [%s]", f.Name())
					}

					t.Endian = oriend
				}
			}

			if f.rdn != nil {
				if e := t.VM.Exec(f.rdn); e != nil {
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
