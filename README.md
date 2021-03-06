bstruct
====

fast binary encoder/decoder in golang with scripts.

its main target is struct encoding/decoding, though you can use it for slice or what...

bstruct uses multiple tags to describe and manipulate a field:

```go
type custom struct {
	Int16 int16 `endian:"big" skip:"w" rdm:"func1"`
}

ct := custom{}
btype := bstruct.MustNew(ct)
decoder := bsutrct.NewDecoder()

if e := decoder.Decode(btype, &ct); e != nil {
	panic(e)
}
```

refer to [gowalker](https://gowalker.org/github.com/xhebox/bstruct).

a practical example, refer to [xhebox/ctx-go](https://github.com/xhebox/ctx-go).

# performance

ok, i know you care about it:

```go
type small struct {
	A       uint32  `endian:"big"`
	Test1   [4]byte `endian:"big"`
	B, C, D int16   `endian:"big"`
	Length  int32   `endian:"big"`
	Test2   [4]byte `endian:"big"`
}

type big struct {
	A       uint32     `endian:"big"`
	Test1   [128]int32 `endian:"big"`
	B, C, D int16      `endian:"big"`
	Length  int32      `endian:"big"`
	Test2   [4]byte    `endian:"big"`
}
```

std stands for binary/encoding, time of generate coder&Type is not counted:

```
BenchmarkSmallDecode-4           5000000               244 ns/op              48 B/op          1 allocs/op
BenchmarkBigDecode-4             3000000               564 ns/op              48 B/op          1 allocs/op
BenchmarkSmallEncode-4          10000000               210 ns/op               0 B/op          0 allocs/op
BenchmarkBigEncode-4             2000000               826 ns/op               0 B/op          0 allocs/op

BenchmarkStdSmallDecode-4        2000000               675 ns/op             136 B/op          9 allocs/op
BenchmarkStdBigDecode-4           500000              2737 ns/op             680 B/op          9 allocs/op
BenchmarkStdSmallEncode-4        2000000               874 ns/op             176 B/op         16 allocs/op
BenchmarkStdBigEncode-4           500000              2922 ns/op            1264 B/op         16 allocs/op

BenchmarkApiDecode-4             1000000              1649 ns/op            1240 B/op         15 allocs/op
BenchmarkApiEncode-4             2000000               839 ns/op             400 B/op          5 allocs/op
```

you can get it by running `go test -bench .`

it should be the fatest one despite its advance function(scripts, align/endian/skip wont effect much). of course, it is a universal machine and wont be faster than protobuf or which serialization library.

# endian tag

set endianess. it's possible to change the result by invoking `SetFlag(FieldFlag)`.

child will inherit parents' endianess. value could be "msb" and "big" or "lsb" and "little".

```go
Int16 int16 `endian:"big"`
Int16 int16 `endian:"msb"`

Int16 int16 `endian:"lsb"`
Int16 int16 `endian:"little"`
```

# skip flag

skip read or write this field. it's possible to change the result by invoking `SetFlag(FieldFlag)`.

value should be a random string, which contains "r", "w" or neither.

```go
Int16 int16 `skip:"r"`
Int16 int16 `skip:"w"`
Int16 int16 `skip:"rwdsfsd"`
```

# align flag

basic-type(uint,int,float,complex) only flag.

read `align` bytes, but just use the first `n` bytes. by default align is equal to the size of the type. align has a maximum, you can get it by `MaxAlign`.

```go
Int16 int16 `align:"16"`
Int16 int16 `align:"32"`
Int16 int16 `align:"64"`
```

# prog flag

include `rdm, rdn, wtm, wtn, type` flags. use `coder.Register` first to register a function, assign flag a function name, it will be invoked. type is talked next section. all function will receive `(interface of root, interface of struct where the current field lies in if possible)`.

```go
Int16 int16 `rdm:"func1"` // read pre
Int16 int16 `rdn:"func2"` // read post
Int16 int16 `wtm:"func3"` // write pre
Int16 int16 `wtn:"func4"` // write post
```

# type flag

you are able to cast type. it's using `reflect.Set` magic, so it's your own duty to avoid `panic`, e.g. type must be of same kind. you can register your own Type by `RegisterType(...)`. specifically, `'some type'` is a syntax sugar to return an constant string.

```go
Int16 int16 `type:"'int8'"` // it's OK
Int16 int16 `type:"'uint8'"` // will panic
Int16 interface{} `type:"'int16'"` // it's OK
```

interface must have a type program. a custom type with `Invalid` kind that implements `CustomRW` will execute their own read/write function.

# varint

bstruct supports varint, too. align makes no effects to varint, but endian does. lsb will decode the first 7-bit group as the right most 7-bit. msb will decode the first 7-bit group as the left most 7-bit.

varint is excepted to use with int types, while uvarint is excepted to use with uint types.

```go
Int16 int32 `type:"'varint'"`
Int16 uint32 `type:"'uvarint'"`
```

# slice

in bstruct, special tags apply to slice.

for convenience, slice has three **reading** modes:

- modelen: mark tag 'length' as a program. prog should return the slice length before reading. non-positive length skip. and if elem is of basic type, it offers a high speed. array internally is implemented as a modelen slice.
- modesize: mark tag 'size' as a program. prog should return the space in bytes before readning. non-positive skip. and if elem is of basic type, it becomes modelen. or it's modeeof with a limitedreader.
- modeeof: it will read until EOF, EOF is not an error. if reader is of type bytes.Reader or bytes.Buffer, while its child is of fix-sized, it becomes modelen.

slice defaults to modeeof, if its child is fix-sized, it goes to modelen eventually.

```go
Int16 []int16
Int16 []int16 `length:"func1"`
Int16 []int16 `size:"func2"` // fallbacks to modelen
Int16 []struct{
	A [4]byte
} `size:"func3"` // but not this one, so more space is taken
```

Note that, slice of struct is always very slow, whatever mode it is at. This is basically a result of trying to support scripts.

# string

as string is immutable, it does not implement slice mode. the only way to change its value is assignment.

# other things

- endian: defaults to host endianess.
- nested slice/string: not allowed. but you can wrap it as a struct, then it's ok.
- since bstruct is using read frequently, i'd recommend bufio if it's a stream. otherwise bytes.Reader is good.
