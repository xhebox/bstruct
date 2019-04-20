package bstruct

// it's API compatible with reflect.Kind(). so it's easier for me to write, and for you to use with reflect.
//
// special cases:
//
// 1. interface{}, int, uint. no one will like unspecific-sized variable when dealing with binary data. these three types must have a type cast program
//
// 2. map. it would be nice to support map, but it's too complicated while you can just unpack binary data simply with a struct{string, uint64} or so.
//
// 3. complex. complex type is too rare, just use float, or a struct of two floats.
//
// 4. pointer. i just can not find a usage of ptr.
type Kind uint

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	uintptrf // wasted
	Float32
	Float64
	complex64f
	complex128f
	Array
	chanf // wasted
	funcf // wasted
	Interface
	mapf // wasted
	ptrf // wasted
	Slice
	stringf // wasted
	Struct
	unsafepointerf // wasted
	UVarint
	Varint
)

func (this Kind) String() string {
	switch this {
	case Invalid:
		return "invalid"
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Int8:
		return "int8"
	case Int16:
		return "int16"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Uint:
		return "uint"
	case Uint8:
		return "uint8"
	case Uint16:
		return "uint16"
	case Uint32:
		return "uint32"
	case Uint64:
		return "uint64"
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case Slice:
		return "slice"
	case Struct:
		return "struct"
	case Interface:
		return "interface{}"
	case UVarint:
		return "uvarint"
	case Varint:
		return "varint"
	default:
		return "wasted"
	}
}

func (this Kind) IsBasic() bool {
	switch this {
	case Bool, Int, Int8, Int16, Int32, Int64, Uint8, Uint16, Uint32, Uint64, Uint, Float32, Float64:
		return true
	default:
		return false
	}
}

// only return size for basic kinds, 0 otherwise
func (this Kind) Size() int {
	switch this {
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
