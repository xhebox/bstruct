package bstruct

import (
	"io"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/xhebox/bstruct/byteorder"
)

var (
	encbuf = make([]byte, MaxAlign)
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
	Runner *Runner
	root   interface{}
}

// just like New() *Type, always create encoder by this function
func NewEncoder() *Encoder {
	return &Encoder{
		Endian: HostEndian,
		Runner: &Runner{
			progs: map[string]func(...interface{}) interface{}{},
		},
	}
}

// pass the generated *Type, and a pointer to data
func (t *Encoder) Encode(w *Type, data interface{}) error {
	t.root = data
	return t.encode(w, w.kind.Size(), reflect.Indirect(reflect.ValueOf(data)))
}

func (t *Encoder) encode(w *Type, align int, v reflect.Value) error {
	switch w.kind {
	case Invalid:
		if v.CanInterface() {
			n, ok := v.Interface().(CustomW)
			if ok {
				return n.Write(t.Wt, t.Endian)
			}
		}
	case UVarint:
		e := byteorder.PutUVarint(t.Wt, t.Endian, v.Uint())
		if e != nil {
			return errors.Wrapf(e, "can not write uvarint")
		}
	case Varint:
		e := byteorder.PutVarint(t.Wt, t.Endian, v.Int())
		if e != nil {
			return errors.Wrapf(e, "can not write varint")
		}
	case Bool, Int8, Int16, Int32, Int64, Uint8, Uint16, Uint32, Uint64, Float32, Float64:
		size := w.kind.Size()

		hdr := reflect.SliceHeader{
			Data: v.UnsafeAddr(),
			Len:  size,
			Cap:  size,
		}

		buf := *(*[]byte)(unsafe.Pointer(&hdr))

		if HostEndian != t.Endian {
			byteorder.ReverseBytes(buf)
		}

		if _, e := t.Wt.Write(buf); e != nil {
			return errors.Wrapf(e, "can not write")
		}

		if HostEndian != t.Endian {
			byteorder.ReverseBytes(buf)
		}

		if align > size {
			if _, e := t.Wt.Write(encbuf[:align-size]); e != nil {
				return errors.Wrapf(e, "can not write")
			}
		}
	case Array, Slice:
		l := v.Len()
		elem := w.slice_elem
		kind := elem.kind

		if kind.IsBasic() {
			size := kind.Size()

			hdr := reflect.SliceHeader{
				Len: l * size,
				Cap: l * size,
			}

			if w.kind == Array {
				hdr.Data = v.UnsafeAddr()
			} else {
				o := (*reflect.SliceHeader)(unsafe.Pointer(v.UnsafeAddr()))
				hdr.Data = o.Data
			}
			buf := *(*[]byte)(unsafe.Pointer(&hdr))

			if HostEndian != t.Endian {
				byteorder.ReverseBuf(buf, size)
			}

			if _, e := t.Wt.Write(buf); e != nil {
				return errors.WithStack(e)
			}

			if HostEndian != t.Endian {
				byteorder.ReverseBuf(buf, size)
			}
		} else {
			for cnt := 0; cnt < l; cnt++ {
				if e := t.encode(elem, align, v.Index(cnt)); e != nil {
					return errors.Wrapf(e, "can not execute encode for elem[%d]", cnt)
				}
			}
		}
	case Struct:
		var vi interface{}

		for k, f := range w.struct_elem {
			var fw = f.rtype
			var fv = v.Field(k)

			if len(f.typ)+len(f.wtm)+len(f.wtn)+len(fw.slice_extra) > 0 {
				vi = v.Interface()
			}

			if l := len(f.typ); l != 0 {
				typ := ""

				if f.typ[0] == '\'' && f.typ[l-1] == '\'' {
					typ = f.typ[1:l]
				} else {
					var ok bool
					typ, ok = t.Runner.exec(f.typ, t.root, vi).(string)
					if !ok {
						return errors.Errorf("can not execute type program")
					}
				}

				rtype, ok := Types[typ]
				if !ok {
					return errors.New("can not resolve type casting")
				}

				fw = rtype
				f.align = rtype.kind.Size()
			}

			if len(f.wtm) != 0 {
				e, ok := t.Runner.exec(f.wtm, t.root, vi).(error)
				if ok {
					return errors.Errorf("can not execute wtm program: %+v", e)
				}
			}

			if f.flag&FlagSkipw == 0 {
				var oriend = t.Endian

				if f.flag&FlagCusEnd != 0 {
					if f.flag&FlagBig != 0 {
						t.Endian = byteorder.BigEndian
					} else {
						t.Endian = byteorder.LittleEndian
					}
				}

				if e := t.encode(fw, f.align, fv); e != nil {
					return errors.Wrapf(e, "can not execute encode for field [%s]", f.Name())
				}

				t.Endian = oriend
			}

			if len(f.wtn) != 0 {
				e, ok := t.Runner.exec(f.wtn, t.root, vi).(error)
				if ok {
					return errors.Errorf("can not execute wtn program: %+v", e)
				}
			}
		}
	default:
		return errors.Errorf("unsupported type: %v\n", w.kind)
	}

	return nil
}
