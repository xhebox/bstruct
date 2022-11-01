package bstruct

import (
	"fmt"
	"go/ast"
	"go/token"
	"unicode"
)

var (
	writerType = &ast.UnaryExpr{X: newSel("bstruct", "Writer"), Op: token.MUL}
	readerType = &ast.UnaryExpr{X: newSel("bstruct", "Reader"), Op: token.MUL}
)

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}

func newMul(l, r ast.Expr) *ast.BinaryExpr {
	return &ast.BinaryExpr{X: l, Y: r, Op: token.MUL}
}

func newSel(l any, r string) *ast.SelectorExpr {
	return &ast.SelectorExpr{X: newIdent(l), Sel: ast.NewIdent(r)}
}

func newPtr(l ast.Expr) *ast.UnaryExpr {
	return &ast.UnaryExpr{X: l, Op: token.AND}
}

func newOpt(l string) string {
	return fmt.Sprintf("__%s", l)
}

func newIdx(l ast.Expr, d any) *ast.IndexExpr {
	switch v := d.(type) {
	case int:
		return &ast.IndexExpr{X: l, Index: intLit(v)}
	case uint:
		return &ast.IndexExpr{X: l, Index: intLit(v)}
	case ast.Expr, string:
		return &ast.IndexExpr{X: l, Index: newIdent(v)}
	default:
		panic("wth")
	}
}

func newLen(l ast.Expr) *ast.CallExpr {
	return newCall("len", l)
}

func newCall(f any, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  newIdent(f),
		Args: args,
	}
}

func newDef(l, r any) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{newIdent(l)},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{newIdent(r)},
	}
}

func newAssign(l, r any) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{newIdent(l)},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{newIdent(r)},
	}
}

func newCallST(f any, args ...ast.Expr) *ast.ExprStmt {
	return &ast.ExprStmt{X: newCall(f, args...)}
}

func unsafePtr(e ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  newSel("unsafe", "Pointer"),
		Args: []ast.Expr{e},
	}
}

type integer interface {
	~int | uint
}

func intLit[T integer](d T) *ast.BasicLit {
	return &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", d)}
}

func newIdent(l any) ast.Expr {
	switch v := l.(type) {
	case string:
		return ast.NewIdent(v)
	case ast.Expr:
		return v
	default:
		panic("wrong")
	}
}
