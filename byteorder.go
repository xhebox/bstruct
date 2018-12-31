package bstruct

import (
	"github.com/xhebox/bstruct/byteorder"
)

var (
	// it's set when one of New(), NewDecoder(), NewEncoder() is called
	HostEndian byteorder.ByteOrder
)

const (
	// false stands for msb
	// true stands for lsb
	BigEndian    = byteorder.ByteOrder(false)
	LittleEndian = byteorder.ByteOrder(true)
)
