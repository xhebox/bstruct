package bstruct

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/xhebox/bstruct/byteorder"
)

type SliceMode uint8

const (
	SliceModeLen SliceMode = iota
	SliceModeSize
	SliceModeEOF
)

var (
	decbuf = make([]byte, MaxAlign)
)

// Decoder has three exported field.
//
// 1. Rd: every time you want to read something new, you need to refresh the reader.
//
// 2. Endian: when you want to change the default Endian on the fly
//
// 3. Runner: hold all callback functions
type Decoder struct {
	Rd     io.Reader
	Endian byteorder.ByteOrder
	Runner *Runner
	root   interface{}
}

// just like New() *Type, always create decoder by this function
func NewDecoder() *Decoder {
	return &Decoder{
		Endian: HostEndian,
		Runner: &Runner{
			progs: map[string]func(...interface{}) interface{}{},
		},
	}
}

// pass the generated *Type, and a pointer to data
func (t *Decoder) Decode(w *Type, data interface{}) error {
	t.root = data
	return t.decode(w, w.kind.Size(), reflect.Indirect(reflect.ValueOf(data)), nil)
}

func (t *Decoder) decode(w *Type, align int, v reflect.Value, pvi interface{}) error {
	switch w.kind {
	case Invalid:
		if v.CanInterface() {
			n, ok := v.Interface().(CustomR)
			if ok {
				return n.Read(t.Rd, t.Endian)
			}
		}
	case UVarint:
		n, e := byteorder.UVarint(t.Rd, t.Endian)
		if e != nil {
			return errors.Wrapf(e, "can not read uvarint")
		}

		v.SetUint(n)
	case Varint:
		n, e := byteorder.Varint(t.Rd, t.Endian)
		if e != nil {
			return errors.Wrapf(e, "can not read varint")
		}

		v.SetInt(n)
	case Bool, Int8, Int16, Int32, Int64, Uint8, Uint16, Uint32, Uint64, Float32, Float64:
		size := w.kind.Size()

		hdr := reflect.SliceHeader{
			Data: v.UnsafeAddr(),
			Len:  size,
			Cap:  size,
		}

		buf := *(*[]byte)(unsafe.Pointer(&hdr))

		if _, e := t.Rd.Read(buf); e != nil {
			return errors.Wrapf(e, "can not read")
		}

		if align > size {
			if _, e := t.Rd.Read(decbuf[:align-size]); e != nil {
				return errors.Wrapf(e, "can not read")
			}
		}

		if HostEndian != t.Endian {
			byteorder.ReverseBytes(buf)
		}
	case Array, Slice:
		var ord = t.Rd
		var mode = w.slice_mode

		switch mode {
		case SliceModeLen:
			if len(w.slice_extra) == 0 {
				break
			}

			var l int
			var ok bool
			if pvi != nil {
				l, ok = t.Runner.exec(w.slice_extra, t.root, pvi).(int)
			} else {
				l, ok = t.Runner.exec(w.slice_extra, t.root).(int)
			}
			if !ok {
				return errors.Errorf("can not execute length program")
			}

			if l > 0 {
				v.Set(reflect.MakeSlice(v.Type(), l, l))
			} else {
				return nil
			}
		case SliceModeSize:
			if len(w.slice_extra) == 0 {
				panic("internal error")
			}

			var l int
			var ok bool
			if pvi != nil {
				l, ok = t.Runner.exec(w.slice_extra, t.root, pvi).(int)
			} else {
				l, ok = t.Runner.exec(w.slice_extra, t.root).(int)
			}
			if !ok {
				return errors.Errorf("can not execute size program")
			}

			if l > 0 {
				sz := w.slice_elem.Kind().Size()
				if sz == 0 {
					t.Rd = io.LimitReader(ord, int64(l))
					mode = SliceModeEOF
				} else {
					cnt := l / sz
					v.Set(reflect.MakeSlice(v.Type(), cnt, cnt))
					mode = SliceModeLen
				}
			} else {
				return nil
			}
		case SliceModeEOF:
			size := w.slice_elem.Size(v.Type().Elem())

			if size == 0 {
				break
			}

			switch n := t.Rd.(type) {
			case *bytes.Buffer:
				cnt := n.Len() / size
				v.Set(reflect.MakeSlice(v.Type(), cnt, cnt))
				mode = SliceModeLen
			case *bytes.Reader:
				cnt := n.Len() / size
				v.Set(reflect.MakeSlice(v.Type(), cnt, cnt))
				mode = SliceModeLen
			}
		}

		switch mode {
		case SliceModeLen:
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

				if _, e := t.Rd.Read(buf); e != nil {
					return errors.WithStack(e)
				}

				if HostEndian != t.Endian {
					byteorder.ReverseBuf(buf, size)
				}
			} else {
				for cnt := 0; cnt < l; cnt++ {
					if e := t.decode(elem, 0, v.Index(cnt), nil); e != nil {
						return errors.Wrapf(e, "can not execute decode for elem[%d]", cnt)
					}
				}
			}
		case SliceModeEOF:
			valtype := v.Type().Elem()
			elem := w.slice_elem
			array := []reflect.Value{}

			for {
				val := reflect.New(valtype).Elem()

				if e := t.decode(elem, 0, val, nil); e != nil {
					t.Rd = ord

					if strings.HasSuffix(e.Error(), "EOF") {
						return nil
					}

					return errors.Wrapf(e, "can not decode elem")
				}

				array = append(array, val)
			}

			v.Set(reflect.Append(v, array...))
		default:
			panic("internal error")
		}

		t.Rd = ord
	case Struct:
		var vi interface{}

		for k, f := range w.struct_elem {
			var fw = f.rtype
			var fv = v.Field(k)

			if len(f.typ)+len(f.rdm)+len(f.rdn)+len(fw.slice_extra) > 0 {
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

			if len(f.rdm) != 0 {
				e, ok := t.Runner.exec(f.rdm, t.root, vi).(error)
				if ok {
					return errors.Errorf("can not execute rdm program: %+v", e)
				}
			}

			if f.flag&FlagSkipr == 0 {
				var oriend = t.Endian

				if f.flag&FlagCusEnd != 0 {
					if f.flag&FlagBig != 0 {
						t.Endian = byteorder.BigEndian
					} else {
						t.Endian = byteorder.LittleEndian
					}
				}

				if e := t.decode(fw, f.align, fv, vi); e != nil {
					return errors.Wrapf(e, "can not execute decode for field [%s]", f.Name())
				}

				t.Endian = oriend
			}

			if len(f.rdn) != 0 {
				e, ok := t.Runner.exec(f.rdn, t.root, vi).(error)
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
