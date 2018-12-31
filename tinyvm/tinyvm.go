package tinyvm

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
	"github.com/robertkrimen/otto/token"
	"github.com/xhebox/bstruct/byteorder"
)

const (
	PUSH0 byte = iota
	PUSH1
	PUSHC
	PUSHR
	PUSHK
	PUSHV
	PUSHENV
	POP
	ADD
	SUB
	MUL
	DIV
	MOD
	SHL
	SHR
	BAND
	BOR
	BXOR
	AND
	OR
	NEG
	NOT
	BNOT
	EQ
	NE
	LT
	LE
	GT
	GE
	IDX
	FLD
	JNC
	JMP
	CALL
	RET
	ASSIGN
)

const (
	vint byte = iota
	vfloat
	vstr
	vref
)

// it's only exported to use with VM.Ret()
type Value struct {
	kind byte
	v1   int64
	v2   float64
	v3   string
	v4   reflect.Value
}

func (t Value) ToInteger() int64 {
	switch t.kind {
	case vint:
		return t.v1
	case vfloat:
		return int64(t.v2)
	case vstr:
		return int64(len(t.v3))
	case vref:
		switch t.v4.Kind() {
		case reflect.Bool:
			if t.v4.Bool() {
				return 1
			} else {
				return 0
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return t.v4.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int64(t.v4.Uint())
		default:
			return 0
		}
	default:
		panic("unknown type")
	}
}

func (t Value) ToString() string {
	switch t.kind {
	case vint:
		return strconv.Itoa(int(t.v1))
	case vfloat:
		return fmt.Sprint(t.v2)
	case vstr:
		return t.v3
	case vref:
		return fmt.Sprint(t.v4)
	default:
		panic("unknown type")
	}
}

func (t Value) ToFloat() float64 {
	switch t.kind {
	case vint:
		return float64(t.v1)
	case vfloat:
		return t.v2
	case vstr:
		return float64(len(t.v3))
	case vref:
		switch t.v4.Kind() {
		case reflect.Float32, reflect.Float64:
			return t.v4.Float()
		case reflect.Bool:
			if t.v4.Bool() {
				return 1
			} else {
				return 0
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(t.v4.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(t.v4.Uint())
		default:
			return 0
		}
	default:
		panic("unknown type")
	}
}

func (t Value) ToGo() reflect.Value {
	switch t.kind {
	case vint:
		return reflect.ValueOf(t.v1)
	case vfloat:
		return reflect.ValueOf(t.v2)
	case vstr:
		return reflect.ValueOf(t.v3)
	case vref:
		switch t.v4.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return reflect.ValueOf(t.v4.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return reflect.ValueOf(int64(t.v4.Uint()))
		case reflect.Float32, reflect.Float64:
			return reflect.ValueOf(t.v4.Float())
		default:
			return t.v4
		}
	default:
		panic("unknown type")
	}
}

func (t Value) Set(v Value) {
	switch t.kind {
	case vint:
		t.v1 = v.ToInteger()
	case vfloat:
		t.v2 = v.ToFloat()
	case vstr:
		t.v3 = v.ToString()
	case vref:
		t.v4.Set(v.ToGo().Convert(t.v4.Type()))
	default:
		panic("unknown type")
	}
}

// a comipler for scripts, it's exported for convenience, same to VM and prog
type Compiler struct {
	Endian byteorder.ByteOrder
	// scripts are able to directly, at bytecode level, access two reflect value 'current' and 'root', and that leads to a need of reflect.Type info if program used these two identifiers.
	C, R, r reflect.Type
	csts    []Value
	buf     []byte
}

func NewCompiler() *Compiler {
	return &Compiler{
		buf: make([]byte, 4),
	}
}

func (t *Compiler) MustCompile(src string) *Prog {
	prog, e := t.Compile(src)
	if e != nil {
		panic(e)
	}

	return prog
}

func (t *Compiler) Compile(src string) (*Prog, error) {
	prog, e := parser.ParseFile(nil, "", src, 0)
	if e != nil {
		return nil, errors.WithStack(e)
	}

	wt := &bytes.Buffer{}
	ret := &Prog{}
	t.csts = t.csts[0:0]

	for _, v := range prog.Body {
		wt.Reset()
		if e := t.icompile(wt, v); e != nil {
			return ret, errors.WithStack(e)
		}
		ret.code = append(ret.code, wt.Bytes()...)
	}

	ret.csts = append(ret.csts, t.csts...)

	return ret, nil
}

func (t *Compiler) addfloat(f float64) uint8 {
	for k, v := range t.csts {
		if v.kind != vstr && v.ToFloat() == f {
			return uint8(k)
		}
	}

	t.csts = append(t.csts, Value{kind: vfloat, v2: f})
	return uint8(len(t.csts)) - 1
}

func (t *Compiler) addint(i int64) uint8 {
	for k, v := range t.csts {
		if v.kind != vstr && v.ToInteger() == i {
			return uint8(k)
		}
	}

	t.csts = append(t.csts, Value{kind: vint, v1: i})
	return uint8(len(t.csts)) - 1
}

func (t *Compiler) addstr(s string) uint8 {
	for k, v := range t.csts {
		if v.kind == vstr && v.ToString() == s {
			return uint8(k)
		}
	}

	t.csts = append(t.csts, Value{kind: vstr, v3: s})
	return uint8(len(t.csts)) - 1
}

func (t *Compiler) icompile(wt io.Writer, node ast.Node) error {
	switch n := node.(type) {
	case nil, *ast.EmptyExpression, *ast.EmptyStatement:
		t.buf[0] = PUSH0
		wt.Write(t.buf[:1])
	case *ast.Identifier:
		switch n.Name {
		case "k":
			t.buf[0] = PUSHK
			wt.Write(t.buf[:1])
		case "root":
			t.r = t.R
			t.buf[0] = PUSHR
			wt.Write(t.buf[:1])
		case "current":
			t.r = t.C
			t.buf[0] = PUSHC
			wt.Write(t.buf[:1])
		default:
			t.r = nil
			t.buf[0] = PUSHENV
			t.buf[1] = t.addstr(n.Name)
			wt.Write(t.buf[:2])
		}
	case *ast.BooleanLiteral:
		if n.Value {
			t.buf[0] = PUSH1
		} else {
			t.buf[0] = PUSH0
		}
		wt.Write(t.buf[:1])
	case *ast.StringLiteral:
		t.buf[0] = PUSHV
		t.buf[1] = t.addstr(n.Value)
		wt.Write(t.buf[:2])
	case *ast.NumberLiteral:
		t.buf[0] = PUSHV
		switch m := n.Value.(type) {
		case int64:
			t.buf[1] = t.addint(m)
		case float64:
			t.buf[1] = t.addfloat(m)
		}
		wt.Write(t.buf[:2])
	case *ast.UnaryExpression:
		if e := t.icompile(wt, n.Operand); e != nil {
			return errors.WithStack(e)
		}

		switch n.Operator {
		case token.NOT:
			t.buf[0] = NOT
		case token.BITWISE_NOT:
			t.buf[0] = BNOT
		case token.EXCLUSIVE_OR:
			t.buf[0] = BNOT
		case token.MINUS:
			t.buf[0] = NEG
		default:
			return errors.Errorf("unsupported unary operator %+v", n)
		}

		wt.Write(t.buf[:1])
	case *ast.BinaryExpression:
		if e := t.icompile(wt, n.Left); e != nil {
			return errors.WithStack(e)
		}

		if e := t.icompile(wt, n.Right); e != nil {
			return errors.WithStack(e)
		}

		switch n.Operator {
		case token.PLUS:
			t.buf[0] = ADD
		case token.MINUS:
			t.buf[0] = SUB
		case token.MULTIPLY:
			t.buf[0] = MUL
		case token.SLASH:
			t.buf[0] = DIV
		case token.REMAINDER:
			t.buf[0] = MOD
		case token.SHIFT_LEFT:
			t.buf[0] = SHL
		case token.SHIFT_RIGHT:
			t.buf[0] = SHR
		case token.AND:
			t.buf[0] = BAND
		case token.OR:
			t.buf[0] = BOR
		case token.EXCLUSIVE_OR:
			t.buf[0] = BXOR
		case token.LOGICAL_AND:
			t.buf[0] = AND
		case token.LOGICAL_OR:
			t.buf[0] = OR
		case token.EQUAL:
			t.buf[0] = EQ
		case token.NOT_EQUAL:
			t.buf[0] = NE
		case token.LESS:
			t.buf[0] = LT
		case token.LESS_OR_EQUAL:
			t.buf[0] = LE
		case token.GREATER:
			t.buf[0] = GT
		case token.GREATER_OR_EQUAL:
			t.buf[0] = GE
		default:
			return errors.Errorf("unsupported binary operator %+v", n)
		}

		wt.Write(t.buf[:1])
	case *ast.SequenceExpression:
		last := len(n.Sequence) - 1
		for _, v := range n.Sequence[:last] {
			if e := t.icompile(wt, v); e != nil {
				return errors.WithStack(e)
			}
			t.buf[0] = POP
			wt.Write(t.buf[:1])
		}

		if e := t.icompile(wt, n.Sequence[last]); e != nil {
			return errors.WithStack(e)
		}
	case *ast.ExpressionStatement:
		if e := t.icompile(wt, n.Expression); e != nil {
			return errors.WithStack(e)
		}

		t.buf[0] = POP
		wt.Write(t.buf[:1])
	case *ast.BlockStatement:
		for _, v := range n.List {
			if e := t.icompile(wt, v); e != nil {
				return errors.WithStack(e)
			}
		}
	case *ast.DotExpression:
		if e := t.icompile(wt, n.Left); e != nil {
			return errors.WithStack(e)
		}

		if t.r == nil || t.r.Kind() != reflect.Struct {
			return errors.Errorf("only root/current and variables from them are able to use dot and bracket %v.%s", t.r, n.Identifier.Name)
		}

		if f, ok := t.r.FieldByName(n.Identifier.Name); ok {
			var v = f.Index[len(f.Index)-1]

			t.buf[0] = PUSHV
			t.buf[1] = t.addint(int64(v))
			wt.Write(t.buf[:2])

			t.r = f.Type
		} else {
			return errors.Errorf("[%s] is not a field of [%v]", n.Identifier.Name, t.r)
		}

		t.buf[0] = FLD
		wt.Write(t.buf[:1])
	case *ast.BracketExpression:
		if e := t.icompile(wt, n.Left); e != nil {
			return errors.WithStack(e)
		}

		if t.r == nil || (t.r.Kind() != reflect.Array && t.r.Kind() != reflect.Slice) {
			return errors.Errorf("only array/slice are able to use bracket %v[%v]", t.r, n.Member)
		}

		if e := t.icompile(wt, n.Member); e != nil {
			return errors.WithStack(e)
		}

		t.buf[0] = IDX
		wt.Write(t.buf[:1])
		t.r = t.r.Elem()
	case *ast.AssignExpression:
		if e := t.icompile(wt, n.Left); e != nil {
			return errors.WithStack(e)
		}

		if e := t.icompile(wt, n.Right); e != nil {
			return errors.WithStack(e)
		}

		t.buf[0] = ASSIGN
		wt.Write(t.buf[:1])
	case *ast.IfStatement:
		var twt = &bytes.Buffer{}
		if e := t.icompile(twt, n.Test); e != nil {
			return errors.WithStack(e)
		}

		var cwt = &bytes.Buffer{}
		if e := t.icompile(cwt, n.Consequent); e != nil {
			return errors.WithStack(e)
		}

		var awt = &bytes.Buffer{}
		if e := t.icompile(awt, n.Alternate); e != nil {
			return errors.WithStack(e)
		}

		// test variable
		wt.Write(twt.Bytes())

		// jump condition
		t.buf[0] = JNC
		t.Endian.PutUint16(t.buf[1:], uint16(cwt.Len())+3)
		wt.Write(t.buf[:3])

		// consequent + 3
		wt.Write(cwt.Bytes())

		t.buf[0] = JMP
		t.Endian.PutUint16(t.buf[1:], uint16(awt.Len()))
		wt.Write(t.buf[:3])

		// alternate
		wt.Write(awt.Bytes())
	case *ast.ConditionalExpression:
		var twt = &bytes.Buffer{}
		if e := t.icompile(twt, n.Test); e != nil {
			return errors.WithStack(e)
		}

		var cwt = &bytes.Buffer{}
		if e := t.icompile(cwt, n.Consequent); e != nil {
			return errors.WithStack(e)
		}

		var awt = &bytes.Buffer{}
		if e := t.icompile(awt, n.Alternate); e != nil {
			return errors.WithStack(e)
		}

		// test variable
		wt.Write(twt.Bytes())

		// jump condition
		t.buf[0] = JNC
		t.Endian.PutUint16(t.buf[1:], uint16(cwt.Len())+3)
		wt.Write(t.buf[:3])

		// consequent + 3
		wt.Write(cwt.Bytes())

		t.buf[0] = JMP
		t.Endian.PutUint16(t.buf[1:], uint16(awt.Len()))
		wt.Write(t.buf[:3])

		// alternate
		wt.Write(awt.Bytes())
	case *ast.CallExpression:
		for _, v := range n.ArgumentList {
			if e := t.icompile(wt, v); e != nil {
				return errors.WithStack(e)
			}
		}

		if e := t.icompile(wt, n.Callee); e != nil {
			return errors.WithStack(e)
		}

		t.buf[0] = CALL
		t.buf[1] = RET
		t.buf[2] = uint8(len(n.ArgumentList))
		wt.Write(t.buf[:3])
	default:
		return errors.Errorf("unsupported ast %+v", n)
	}

	return nil
}

type Prog struct {
	code []byte
	csts []Value
}

type VM struct {
	sidx   int
	Endian byteorder.ByteOrder
	// k is also reserved like 'root'&'current', VM needs its pointer
	K int64
	// you need to pass reflect.Value if program used 'root' and 'current' identifiers
	Current, Root reflect.Value
	env           map[string]Value
	stack         []Value
}

// envsz tell you can set how many external variables
// stsz tell the stack size
func (t *VM) Init(envsz, stsz int) {
	t.env = make(map[string]Value, stsz)
	t.stack = make([]Value, envsz)
}

func (t *VM) push(x Value) {
	t.stack[t.sidx] = x
	t.sidx++
}

func (t *VM) pop() {
	t.sidx--
}

func (t *VM) Exec(prog *Prog) error {
	var oc = prog.code

	for c, e := 0, len(oc); c < e; {
		switch oc[c] {
		case PUSH0:
			t.push(Value{kind: vint, v1: 0})
			c++
		case PUSH1:
			t.push(Value{kind: vint, v1: 1})
			c++
		case PUSHC:
			t.push(Value{kind: vref, v4: t.Current})
			c++
		case PUSHK:
			t.push(Value{kind: vint, v1: t.K})
			c++
		case PUSHR:
			t.push(Value{kind: vref, v4: t.Root})
			c++
		case PUSHV:
			t.push(prog.csts[oc[c+1]])
			c += 2
		case PUSHENV:
			if v, ok := t.env[prog.csts[oc[c+1]].ToString()]; ok {
				t.push(v)
			} else {
				return errors.Errorf("failed to resolv [%s]", prog.csts[oc[c+1]].ToString())
			}
			c += 2
		case POP:
			t.pop()
			c++
		case ADD, SUB, MUL, DIV, MOD, SHL, SHR, BAND, BOR, BXOR, AND, OR, EQ, NE, LT, LE, GT, GE:
			switch t.stack[t.sidx-2].kind {
			case vint:
				switch t.stack[t.sidx-1].kind {
				case vfloat:
					t.stack[t.sidx-2].v2 = t.stack[t.sidx-2].ToFloat()
					goto bfloat
				case vint:
					goto bint
				default:
					t.stack[t.sidx-1].v1 = t.stack[t.sidx-1].ToInteger()
					goto bint
				}
			case vfloat:
				if t.stack[t.sidx-1].kind != vfloat {
					t.stack[t.sidx-1].v2 = t.stack[t.sidx-1].ToFloat()
				}

				goto bfloat
			case vstr:
				switch t.stack[t.sidx-1].kind {
				case vint:
					t.stack[t.sidx-2].v1 = t.stack[t.sidx-2].ToInteger()
					goto bint
				case vfloat:
					t.stack[t.sidx-2].v2 = t.stack[t.sidx-2].ToFloat()
					goto bfloat
				case vstr:
					goto bstr
				default:
					t.stack[t.sidx-1].v3 = t.stack[t.sidx-1].ToString()
					goto bstr
				}
			case vref:
				switch t.stack[t.sidx-1].kind {
				case vint:
					t.stack[t.sidx-2].v1 = t.stack[t.sidx-2].ToInteger()
					goto bint
				case vfloat:
					t.stack[t.sidx-2].v2 = t.stack[t.sidx-2].ToFloat()
					goto bfloat
				case vstr:
					t.stack[t.sidx-2].v3 = t.stack[t.sidx-2].ToString()
					goto bstr
				case vref:
					t.stack[t.sidx-2].v1 = t.stack[t.sidx-2].ToInteger()
					t.stack[t.sidx-1].v1 = t.stack[t.sidx-1].ToInteger()
					goto bint
				}
			}

		bint:
			t.stack[t.sidx-2].kind = vint
			switch oc[c] {
			case ADD:
				t.stack[t.sidx-2].v1 += t.stack[t.sidx-1].v1
			case SUB:
				t.stack[t.sidx-2].v1 -= t.stack[t.sidx-1].v1
			case MUL:
				t.stack[t.sidx-2].v1 *= t.stack[t.sidx-1].v1
			case DIV:
				t.stack[t.sidx-2].v1 /= t.stack[t.sidx-1].v1
			case MOD:
				t.stack[t.sidx-2].v1 %= t.stack[t.sidx-1].v1
			case SHL:
				t.stack[t.sidx-2].v1 <<= uint64(t.stack[t.sidx-1].v1)
			case SHR:
				t.stack[t.sidx-2].v1 >>= uint64(t.stack[t.sidx-1].v1)
			case BAND:
				t.stack[t.sidx-2].v1 &= t.stack[t.sidx-1].v1
			case BOR:
				t.stack[t.sidx-2].v1 |= t.stack[t.sidx-1].v1
			case BXOR:
				t.stack[t.sidx-2].v1 ^= t.stack[t.sidx-1].v1
			case AND:
				if t.stack[t.sidx-2].v1 != 0 && t.stack[t.sidx-1].v1 != 0 {
					t.stack[t.sidx-2].v1 = 1
				} else {
					t.stack[t.sidx-2].v1 = 0
				}
			case OR:
				if t.stack[t.sidx-2].v1 != 0 || t.stack[t.sidx-1].v1 != 0 {
					t.stack[t.sidx-2].v1 = 1
				} else {
					t.stack[t.sidx-2].v1 = 0
				}
			case EQ:
				if t.stack[t.sidx-2].v1 == t.stack[t.sidx-1].v1 {
					t.stack[t.sidx-2].v1 = 1
				} else {
					t.stack[t.sidx-2].v1 = 0
				}
			case NE:
				if t.stack[t.sidx-2].v1 != t.stack[t.sidx-1].v1 {
					t.stack[t.sidx-2].v1 = 1
				} else {
					t.stack[t.sidx-2].v1 = 0
				}
			case LT:
				if t.stack[t.sidx-2].v1 < t.stack[t.sidx-1].v1 {
					t.stack[t.sidx-2].v1 = 1
				} else {
					t.stack[t.sidx-2].v1 = 0
				}
			case LE:
				if t.stack[t.sidx-2].v1 <= t.stack[t.sidx-1].v1 {
					t.stack[t.sidx-2].v1 = 1
				} else {
					t.stack[t.sidx-2].v1 = 0
				}
			case GT:
				if t.stack[t.sidx-2].v1 > t.stack[t.sidx-1].v1 {
					t.stack[t.sidx-2].v1 = 1
				} else {
					t.stack[t.sidx-2].v1 = 0
				}
			case GE:
				if t.stack[t.sidx-2].v1 >= t.stack[t.sidx-1].v1 {
					t.stack[t.sidx-2].v1 = 1
				} else {
					t.stack[t.sidx-2].v1 = 0
				}
			default:
				t.stack[t.sidx-2].v1 = 0
			}
			t.pop()
			c++
			continue

		bfloat:
			t.stack[t.sidx-2].kind = vfloat
			switch oc[c] {
			case ADD:
				t.stack[t.sidx-2].v2 += t.stack[t.sidx-1].v2
			case SUB:
				t.stack[t.sidx-2].v2 -= t.stack[t.sidx-1].v2
			case MUL:
				t.stack[t.sidx-2].v2 *= t.stack[t.sidx-1].v2
			case DIV:
				t.stack[t.sidx-2].v2 /= t.stack[t.sidx-1].v2
			case BAND:
				t.stack[t.sidx-2].v2 = float64(uint64(t.stack[t.sidx-2].v2) & uint64(t.stack[t.sidx-1].v2))
			case BOR:
				t.stack[t.sidx-2].v2 = float64(uint64(t.stack[t.sidx-2].v2) | uint64(t.stack[t.sidx-1].v2))
			case BXOR:
				t.stack[t.sidx-2].v2 = float64(uint64(t.stack[t.sidx-2].v2) ^ uint64(t.stack[t.sidx-1].v2))
			case EQ:
				if t.stack[t.sidx-2].v2 == t.stack[t.sidx-1].v2 {
					t.stack[t.sidx-2].v2 = 1
				} else {
					t.stack[t.sidx-2].v2 = 0
				}
			case NE:
				if t.stack[t.sidx-2].v2 != t.stack[t.sidx-1].v2 {
					t.stack[t.sidx-2].v2 = 1
				} else {
					t.stack[t.sidx-2].v2 = 0
				}
			case LT:
				if t.stack[t.sidx-2].v2 < t.stack[t.sidx-1].v2 {
					t.stack[t.sidx-2].v2 = 1
				} else {
					t.stack[t.sidx-2].v2 = 0
				}
			case LE:
				if t.stack[t.sidx-2].v2 <= t.stack[t.sidx-1].v2 {
					t.stack[t.sidx-2].v2 = 1
				} else {
					t.stack[t.sidx-2].v2 = 0
				}
			case GT:
				if t.stack[t.sidx-2].v2 > t.stack[t.sidx-1].v2 {
					t.stack[t.sidx-2].v2 = 1
				} else {
					t.stack[t.sidx-2].v2 = 0
				}
			case GE:
				if t.stack[t.sidx-2].v2 >= t.stack[t.sidx-1].v2 {
					t.stack[t.sidx-2].v2 = 1
				} else {
					t.stack[t.sidx-2].v2 = 0
				}
			default:
				t.stack[t.sidx-2].v2 = 0
			}
			t.pop()
			c++
			continue

		bstr:
			t.stack[t.sidx-2].kind = vstr
			switch oc[c] {
			case ADD:
				t.stack[t.sidx-2].v3 = t.stack[t.sidx-2].v3 + t.stack[t.sidx-1].v3
			default:
				t.stack[t.sidx-2].v3 = ""
			}
			t.pop()
			c++
			continue
		case NEG, NOT, BNOT:
			switch t.stack[t.sidx-1].kind {
			case vint:
				switch oc[c] {
				case NEG:
					t.stack[t.sidx-1].v1 = -t.stack[t.sidx-1].v1
				case NOT:
					if t.stack[t.sidx-1].v1 == 0 {
						t.stack[t.sidx-1].v1 = 1
					} else {
						t.stack[t.sidx-1].v1 = 0
					}
				case BNOT:
					t.stack[t.sidx-1].v1 = ^t.stack[t.sidx-1].v1
				default:
					t.stack[t.sidx-1].v1 = 0
				}
			case vfloat:
				switch oc[c] {
				case NEG:
					t.stack[t.sidx-1].v2 = -t.stack[t.sidx-1].v2
				case NOT:
					if t.stack[t.sidx-1].v2 == 0 {
						t.stack[t.sidx-1].v2 = 1
					} else {
						t.stack[t.sidx-1].v2 = 0
					}
				case BNOT:
					t.stack[t.sidx-1].v2 = float64(^uint64(t.stack[t.sidx-1].v2))
				default:
					t.stack[t.sidx-1].v2 = 0
				}
			case vstr:
				t.stack[t.sidx-1].v3 = ""
			case vref:
				t.stack[t.sidx-1].v1 = 0
			}
			c++
		case FLD:
			var x = t.stack[t.sidx-2].v4

			if x.Kind() != reflect.Struct {
				return errors.Errorf("except a struct")
			}

			var idx = int(t.stack[t.sidx-1].v1)

			if idx >= x.NumField() {
				return errors.Errorf("idx is overflow")
			}

			t.stack[t.sidx-2].v4 = x.Field(idx)
			t.pop()
			c++
		case IDX:
			var x = t.stack[t.sidx-2].v4

			if kind := x.Kind(); kind != reflect.Slice && kind != reflect.Array {
				return errors.Errorf("except a slice/array in register b")
			}

			var idx = int(t.stack[t.sidx-1].v1)

			if idx >= x.Len() {
				return errors.Errorf("idx is overflow")
			}

			t.stack[t.sidx-2].v4 = x.Index(idx)
			t.pop()
			c++
		case JNC:
			var jmp = t.stack[t.sidx-1].ToInteger() == 0

			if jmp {
				l := t.Endian.Uint16(oc[c+1:])
				c += 3 + int(l)
			} else {
				c += 3
			}
		case JMP:
			l := t.Endian.Uint16(oc[c+1:])
			c += 3 + int(l)
		case ASSIGN:
			t.stack[t.sidx-2].Set(t.stack[t.sidx-1])
			t.pop()
			c++
		case CALL:
			var n = t.sidx - 1
			var x = t.stack[n].v4
			if x.Kind() != reflect.Func {
				return errors.Errorf("except a function")
			}

			var argN = x.Type().NumIn()
			var args = []reflect.Value{}
			for _, v := range t.stack[n-argN : n] {
				args = append(args, v.ToGo())
			}

			y := x.Call(args)

			if len(y) == 0 {
				t.stack[n-argN] = Value{kind: vint, v1: 0}
			} else if len(y) == 1 {
				t.stack[n-argN] = Value{kind: vref, v4: y[0]}
			} else {
				var r = []interface{}{}

				for _, v := range y {
					r = append(r, v.Interface())
				}

				t.stack[n-argN] = Value{kind: vref, v4: reflect.ValueOf(r)}
			}
			c++
		case RET:
			t.sidx -= int(oc[c+1])
			c += 2
		default:
			return errors.Errorf("unsupported bytecode")
		}
	}
	return nil
}

func (t *VM) Ret() Value {
	return t.stack[t.sidx]
}

func (this *VM) Set(name string, val interface{}) {
	switch n := val.(type) {
	case bool:
		if n {
			this.env[name] = Value{kind: vint, v1: 0}
		} else {
			this.env[name] = Value{kind: vint, v1: 1}
		}
	case int:
		this.env[name] = Value{kind: vint, v1: int64(n)}
	case int8:
		this.env[name] = Value{kind: vint, v1: int64(n)}
	case int16:
		this.env[name] = Value{kind: vint, v1: int64(n)}
	case int32:
		this.env[name] = Value{kind: vint, v1: int64(n)}
	case int64:
		this.env[name] = Value{kind: vint, v1: n}
	case uint:
		this.env[name] = Value{kind: vint, v1: int64(n)}
	case uint8:
		this.env[name] = Value{kind: vint, v1: int64(n)}
	case uint16:
		this.env[name] = Value{kind: vint, v1: int64(n)}
	case uint32:
		this.env[name] = Value{kind: vint, v1: int64(n)}
	case uint64:
		this.env[name] = Value{kind: vint, v1: int64(n)}
	case float32:
		this.env[name] = Value{kind: vfloat, v2: float64(n)}
	case float64:
		this.env[name] = Value{kind: vfloat, v2: n}
	case string:
		this.env[name] = Value{kind: vstr, v3: n}
	default:
		this.env[name] = Value{kind: vref, v4: reflect.ValueOf(n)}
	}
}
