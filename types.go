package bstruct

var types = map[string]*Type{
	"bool":    &Type{kind: Bool},
	"int8":    &Type{kind: Int8},
	"int16":   &Type{kind: Int16},
	"int32":   &Type{kind: Int32},
	"int64":   &Type{kind: Int64},
	"byte":    &Type{kind: Uint8},
	"uint8":   &Type{kind: Uint8},
	"uint16":  &Type{kind: Uint16},
	"uint32":  &Type{kind: Uint32},
	"uint64":  &Type{kind: Uint64},
	"float32": &Type{kind: Float32},
	"float64": &Type{kind: Float64},
}

// register new Type for type cast program
func RegisterType(name string, t *Type) {
	types[name] = t
}