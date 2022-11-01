package bstruct

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"math"
	"unicode"
)

const defWrap = 77

type builtField struct {
	field *Field
	typ   ast.GenDecl
	enc   ast.FuncDecl
	dec   ast.FuncDecl
	extra []ast.Decl
}

type Builder struct {
	cnt      uint
	getter   bool
	setter   bool
	lineWrap int
	imports  *ast.GenDecl
	types    map[string]builtField
}

func NewBuilder() *Builder {
	return &Builder{
		types:    make(map[string]builtField),
		lineWrap: defWrap,
	}
}

func (e *Builder) Getter(f bool) *Builder {
	e.getter = f
	return e
}

func (e *Builder) Setter(f bool) *Builder {
	e.setter = f
	return e
}

func (e *Builder) SetLineWrap(wrap int) *Builder {
	if wrap <= 0 {
		e.lineWrap = defWrap
	}
	return e
}

func (e *Builder) encPrim(writer ast.Expr, ptr ast.Expr, s *Field) (stmts []ast.Stmt) {
	switch {
	case s.typ.IsPrimitive():
		stmts = append(stmts, newCallST(
			newSel(writer, "Copy"),
			unsafePtr(newPtr(ptr)),
			intLit(s.typ.Size()),
		))
	case s.typ.IsType(FieldSlice) || s.typ.IsType(FieldString):
		stmts = append(stmts,
			newCallST(
				newSel(writer, "WriteLen"),
				newLen(ptr),
			),
		)
		if s.sliceType.typ.IsPrimitive() {
			hdr := e.newIdent()
			var bstmts []ast.Stmt
			if s.typ.IsType(FieldString) {
				bstmts = append(bstmts,
					newDef(hdr, newCall(&ast.ParenExpr{X: &ast.UnaryExpr{X: newSel("reflect", "StringHeader"), Op: token.MUL}}, unsafePtr(newPtr(ptr)))),
				)
			} else {
				bstmts = append(bstmts,
					newDef(hdr, newCall(&ast.ParenExpr{X: &ast.UnaryExpr{X: newSel("reflect", "SliceHeader"), Op: token.MUL}}, unsafePtr(newPtr(ptr)))),
				)
			}
			bstmts = append(bstmts,
				newCallST(
					newSel(writer, "Copy"),
					unsafePtr(newSel(hdr, "Data")),
					newMul(intLit(s.sliceType.typ.Size()), newLen(ptr)),
				),
			)
			stmts = append(stmts, &ast.IfStmt{
				Cond: &ast.BinaryExpr{X: newLen(ptr), Op: token.GTR, Y: intLit(0)},
				Body: &ast.BlockStmt{List: bstmts},
			})
		} else {
			i := e.newIdent()
			stmts = append(stmts, &ast.RangeStmt{
				Key:  i,
				Tok:  token.DEFINE,
				X:    ptr,
				Body: &ast.BlockStmt{List: e.encField(writer, newIdx(ptr, i), s.sliceType)},
			})
		}
	case s.typ.IsType(FieldStruct):
		for i := range s.strucFields {
			bstmts := e.encField(writer, newSel(ptr, s.strucFields[i].strucName), s.strucFields[i].Field)
			if s.strucFields[i].optional {
				hasField := newSel(ptr, newOpt(s.strucFields[i].strucName))
				stmts = append(stmts, e.encPrim(writer, hasField, New(FieldBool))...)
				stmts = append(stmts, &ast.IfStmt{
					Cond: hasField,
					Body: &ast.BlockStmt{List: bstmts},
				})
			} else {
				stmts = append(stmts, bstmts...)
			}
		}
	case s.typ.IsType(FieldCustom):
		if s.cusenc != nil {
			stmts = s.cusenc(writer, ptr, s)
		}
	default:
		panic("wth")
	}
	return
}

func (e *Builder) encField(writer ast.Expr, ptr ast.Expr, s *Field) (stmts []ast.Stmt) {
	if _, ok := e.types[s.typename]; ok {
		stmts = append(stmts, newCallST(
			newSel(ptr, "Encode"),
			writer,
		))
		return
	}

	stmts = e.encPrim(writer, ptr, s)
	return
}

func (e *Builder) decPrim(reader ast.Expr, ptr ast.Expr, s *Field) (stmts []ast.Stmt) {
	switch {
	case s.typ.IsPrimitive():
		stmts = append(stmts, newCallST(
			newSel(reader, "Copy"),
			unsafePtr(newPtr(ptr)),
			intLit(s.typ.Size()),
		))
	case s.typ.IsType(FieldSlice) || s.typ.IsType(FieldString):
		length := e.newIdent()
		stmts = append(stmts, newDef(length, newCall(newSel(reader, "ReadLen"))))
		var bstmts []ast.Stmt
		if s.sliceType.typ.IsPrimitive() {
			hdr := e.newIdent()
			if s.typ.IsType(FieldSlice) {
				bstmts = append(bstmts,
					newDef(hdr, newCall(&ast.ParenExpr{X: &ast.UnaryExpr{X: newSel("reflect", "SliceHeader"), Op: token.MUL}}, unsafePtr(newPtr(ptr)))),
				)
			} else {
				bstmts = append(bstmts,
					newDef(hdr, newCall(&ast.ParenExpr{X: &ast.UnaryExpr{X: newSel("reflect", "StringHeader"), Op: token.MUL}}, unsafePtr(newPtr(ptr)))),
				)
			}
			bstmts = append(bstmts,
				newAssign(newSel(hdr, "Len"), length),
				newAssign(newSel(hdr, "Data"), newCall(
					newSel(reader, "Read"),
					newMul(intLit(s.sliceType.typ.Size()), length),
				)),
			)
			if s.typ.IsType(FieldSlice) {
				bstmts = append(bstmts,
					newAssign(newSel(hdr, "Cap"), length),
				)
			}
		} else {
			i := e.newIdent()
			bstmts = append(bstmts,
				newAssign(ptr, newCall("make", &ast.ArrayType{Elt: e.typWrap(s.sliceType)}, length)),
				&ast.RangeStmt{
					Key:  newIdent(i),
					Tok:  token.DEFINE,
					X:    ptr,
					Body: &ast.BlockStmt{List: e.decField(reader, newIdx(ptr, i), s.sliceType)},
				},
			)
		}
		stmts = append(stmts, &ast.IfStmt{
			Cond: &ast.BinaryExpr{X: length, Op: token.GTR, Y: intLit(0)},
			Body: &ast.BlockStmt{List: bstmts},
		})
	case s.typ.IsType(FieldStruct):
		for i := range s.strucFields {
			bstmts := e.decField(reader, newSel(ptr, s.strucFields[i].strucName), s.strucFields[i].Field)
			if s.strucFields[i].optional {
				hasField := newSel(ptr, newOpt(s.strucFields[i].strucName))
				stmts = append(stmts, e.decPrim(reader, hasField, New(FieldBool))...)
				stmts = append(stmts, &ast.IfStmt{
					Cond: hasField,
					Body: &ast.BlockStmt{List: bstmts},
				})
			} else {
				stmts = append(stmts, bstmts...)
			}
		}
	case s.typ.IsType(FieldCustom):
		if s.cusdec != nil {
			stmts = s.cusdec(reader, ptr, s)
		}
	default:
		panic("wth")
	}
	return
}

func (e *Builder) decField(reader, ptr ast.Expr, s *Field) (stmts []ast.Stmt) {
	if _, ok := e.types[s.typename]; ok {
		stmts = append(stmts, newCallST(
			newSel(ptr, "Decode"),
			reader,
		))
		return
	}

	stmts = e.decPrim(reader, ptr, s)
	return
}

func (e *Builder) newIdent() *ast.Ident {
	r := ast.NewIdent(fmt.Sprintf("v%d", e.cnt))
	e.cnt++
	return r
}

func (e *Builder) typPrim(s *Field) ast.Expr {
	switch {
	case s.typ.IsPrimitive() || s.typ.IsType(FieldString):
		return newIdent(s.typ.String())
	case s.typ.IsType(FieldSlice):
		return &ast.ArrayType{
			Elt: e.typWrap(s.sliceType),
		}
	case s.typ.IsType(FieldStruct):
		var fields []*ast.Field
		for _, field := range s.strucFields {
			if field.optional {
				fields = append(fields, &ast.Field{
					Names: []*ast.Ident{ast.NewIdent(newOpt(field.strucName))},
					Type:  newIdent(FieldBool.String()),
				})
			}
			fields = append(fields, &ast.Field{
				Names:   []*ast.Ident{ast.NewIdent(field.strucName)},
				Comment: e.commentGroup(field.comment),
				Type:    e.typWrap(field.Field),
			})
		}
		return &ast.StructType{Fields: &ast.FieldList{List: fields}}
	case s.typ.IsType(FieldCustom):
		return s.custyp
	default:
		return newIdent("invalid")
	}
}

func (e *Builder) typWrap(s *Field) ast.Expr {
	if _, ok := e.types[s.typename]; ok {
		return newIdent(s.typename)
	}

	return e.typPrim(s)
}

func (e *Builder) getFunc(el *Field, name string) (*ast.FuncDecl, ast.Expr) {
	typ := &ast.UnaryExpr{X: newIdent(el.typename), Op: token.MUL}
	idt := ast.NewIdent("v")
	var val ast.Expr = idt
	if t := el.typ; !t.IsType(FieldStruct) {
		val = &ast.ParenExpr{X: &ast.UnaryExpr{X: val, Op: token.MUL}}
	}
	return &ast.FuncDecl{
		Name: ast.NewIdent(name),
		Recv: &ast.FieldList{List: []*ast.Field{
			{Names: []*ast.Ident{idt}, Type: typ},
		}},
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: []*ast.Field{}},
			Results: &ast.FieldList{List: []*ast.Field{}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{}},
	}, val
}

func (e *Builder) getFieldGetter(p *Field, el StructField) ast.Decl {
	typ := e.typPrim(el.Field)

	var getterName string
	if unicode.IsUpper(rune(el.strucName[0])) {
		getterName = fmt.Sprintf("Get%s", el.strucName)
	} else {
		getterName = capitalize(el.strucName)
	}
	getter, val := e.getFunc(p, getterName)
	getter.Type.Results.List = append(getter.Type.Results.List, &ast.Field{Type: typ})
	getter.Body.List = append(getter.Body.List, &ast.ReturnStmt{Results: []ast.Expr{newSel(val, el.strucName)}})

	return getter
}

func (e *Builder) getFieldSetter(p *Field, el StructField) ast.Decl {
	typ := e.typPrim(el.Field)

	setterName := fmt.Sprintf("Set%s", capitalize(el.strucName))
	setter, val := e.getFunc(p, setterName)
	sval := ast.NewIdent("i")
	setter.Type.Params.List = append(setter.Type.Params.List, &ast.Field{Names: []*ast.Ident{sval}, Type: typ})
	setter.Body.List = append(setter.Body.List, newAssign(newSel(val, el.strucName), sval))

	return setter
}

func (e *Builder) Process() {
	e.cnt = 0

	e.imports = &ast.GenDecl{
		Tok: token.IMPORT,
	}
	e.imports.Specs = append(e.imports.Specs,
		&ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"unsafe\"",
			},
		},
		&ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"reflect\"",
			},
		},
		&ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"github.com/xhebox/bstruct\"",
			},
		},
	)

	for name, el := range e.types {
		el.typ = ast.GenDecl{
			Tok: token.TYPE,
			Doc: e.commentGroup(el.field.comment),
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: ast.NewIdent(name),
					Type: e.typPrim(el.field),
				},
			},
		}

		writer := ast.NewIdent("wt")
		enc, val := e.getFunc(el.field, "Encode")
		enc.Type.Params.List = append(enc.Type.Params.List, &ast.Field{Names: []*ast.Ident{writer}, Type: writerType})
		enc.Body.List = append(enc.Body.List, e.encPrim(writer, val, el.field)...)
		el.enc = *enc

		reader := ast.NewIdent("rd")
		dec, val := e.getFunc(el.field, "Decode")
		dec.Type.Params.List = append(dec.Type.Params.List, &ast.Field{Names: []*ast.Ident{reader}, Type: readerType})
		dec.Body.List = append(dec.Body.List, e.decPrim(reader, val, el.field)...)
		el.dec = *dec

		el.extra = e.extraDecl(el.field)

		e.types[name] = el
	}
}

func (e *Builder) extraDecl(el *Field) (decls []ast.Decl) {
	if el.typ.IsType(FieldStruct) && e.getter {
		for _, field := range el.strucFields {
			decls = append(decls, e.getFieldGetter(el, field))
		}
	}

	if el.typ.IsType(FieldStruct) && e.setter {
		for _, field := range el.strucFields {
			decls = append(decls, e.getFieldSetter(el, field))
		}
	}

	return
}

func (e *Builder) Print(buf io.Writer, pak string) error {
	ts := token.NewFileSet()
	cfg := printer.Config{
		Mode:     printer.TabIndent,
		Tabwidth: 2,
	}
	file := &ast.File{
		Name:  ast.NewIdent(pak),
		Decls: []ast.Decl{e.imports},
	}
	for _, e := range e.types {
		el := e
		file.Decls = append(file.Decls, &el.typ, &el.enc, &el.dec)
		file.Decls = append(file.Decls, el.extra...)
	}
	if err := cfg.Fprint(buf, ts, file); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(buf, "\n"); err != nil {
		return err
	}
	return nil
}

func (e *Builder) commentGroup(comment string) *ast.CommentGroup {
	if len(comment) == 0 {
		return nil
	}

	strLength := len(comment)
	splitedLength := int(math.Ceil(float64(strLength) / float64(e.lineWrap)))
	splited := make([]*ast.Comment, splitedLength)
	var start, stop int
	for i := 0; i < splitedLength; i += 1 {
		start = i * e.lineWrap
		stop = start + e.lineWrap
		if stop > strLength {
			stop = strLength
		}
		splited[i] = &ast.Comment{Text: fmt.Sprintf("// %s", comment[start:stop])}
	}
	return &ast.CommentGroup{List: splited}
}
