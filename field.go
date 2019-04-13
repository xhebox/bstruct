package bstruct

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// implements field
// normal field implementation
type Field struct {
	rtype *Type
	name  string
	flag  FieldFlag
	align int
	prog  map[string]string
}

func newField(field reflect.StructField) (*Field, error) {
	var flag FieldFlag
	var align int

	if len(field.Tag) != 0 {
		if end := field.Tag.Get("endian"); len(end) != 0 {
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
			skip := field.Tag.Get("skip")
			if strings.Contains(skip, "r") {
				flag |= FlagSkipr
			}
			if strings.Contains(skip, "w") {
				flag |= FlagSkipw
			}
		}

		if alignstr := field.Tag.Get("align"); len(alignstr) != 0 {
			align, e := strconv.Atoi(alignstr)
			if e != nil {
				return nil, e
			}

			if align > 16 {
				return nil, errors.Errorf("align has an upper limit of 16 bytes")
			}
		}
	}

	return &Field{
		flag:  flag,
		align: align,
		prog:  newProgs(field.Tag),
	}, nil
}

// Field Name
func (this *Field) Name() string {
	return this.name
}

// Field type
func (this *Field) Type() *Type {
	return this.rtype
}

// fieldFlag of the field
func (this *Field) Flag() FieldFlag {
	return this.flag
}

// you're allowed to change fieldFlag on the fly
func (this *Field) SetFlag(flag FieldFlag) {
	this.flag = flag
}

// field align
// field flag only makes sense on basic types or slice of basic types
func (this *Field) Align() int {
	return this.align
}

// align should between [size, 16]
func (this *Field) SetAlign(arg int) {
	sz := basicsize(this.rtype.Kind())
	if arg >= sz && arg <= 16 {
		this.align = arg
	}
}
