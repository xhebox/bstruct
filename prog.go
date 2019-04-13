package bstruct

import (
	"reflect"
)

type progs map[string]string

func newProgs(tag reflect.StructTag) progs {
	return progs{
		"type": tag.Get("tpm"),
		"rdm":  tag.Get("rdm"),
		"rdn":  tag.Get("rdn"),
		"wtm":  tag.Get("wtm"),
		"wtn":  tag.Get("wtn"),
	}
}

type runner struct {
	progs map[string]func(...interface{}) interface{}
}

func (c *runner) Register(name string, f func(...interface{}) interface{}) {
	c.progs[name] = f
}

func (c *runner) exec(name string, s ...interface{}) interface{} {
	f, ok := c.progs[name]
	if !ok {
		return nil
	}

	return f(s...)
}
