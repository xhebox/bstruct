bstruct
====

fast binary encoder/decoder in golang with scripts.

its main target is struct encoding/decoding, though you can use it for slice or what...

bstruct uses multiple tags to describe and manipulate a field:

```go
type custom struct {
	Int16 int16 `endian:"big" skip:"w" prog:"view(root.Int16)"`
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
 A       uint32 `endian:"little"`
 Test1   [4]byte
 B, C, D int16
 Length  int32 `endian:"big"`
 Test2   [4]byte
}

type big struct {
 A       uint32 `endian:"little"`
 Test1   [512]byte
 B, C, D int16
 Length  int32 `endian:"big"`
 Test2   [4]byte
}
```

std stands for binary/encoding:

```
BenchmarkSmallDecode-4           3000000               429 ns/op
BenchmarkStdSmallDecode-4        2000000               667 ns/op
BenchmarkBigDecode-4             3000000               434 ns/op
BenchmarkStdBigDecode-4           200000              6393 ns/op
BenchmarkSmallEncode-4           5000000               380 ns/op
BenchmarkStdSmallEncode-4        2000000               958 ns/op
BenchmarkBigEncode-4             2000000               668 ns/op
BenchmarkStdBigEncode-4           200000              7182 ns/op
```

you can get it by running `go test -bench .`

it should be the fatest one despite its advance function(scripts, align/endian/skip wont effect much). of course, it is a universal machine and wont be faster than protobuf or which serialization library.

the huge difference as data gets larger is a made by a speed hack.

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

read `align` bytes, but just use the first `n` bytes. by default align is equal to the size of the type. align has a maximum exactly same as the internal buffer, which is currently 128 bytes.

```go
Int16 int16 `align:"16"`
Int16 int16 `align:"32"`
Int16 int16 `align:"64"`
```

# type flag

you are able to cast type. it's using `reflect.Set` magic, so it's your own duty to avoid `panic`, e.g. type must be of same kind. you can register your own Type by `RegisterType(...)`.

```go
Int16 int16 `type:"int8"` // it's OK
Int16 int16 `type:"uint8"` // will panic
Int16 interface{} `type:"int16"` // it's OK
```

interface must have a type program.

# prog flag

the program, must compile, in hostEndian.

it's like a super subset of C, only basic operator, assign, call, if statement and bool, int64, float64, string. there's no support for loop, variable/function definition, macro, etc...

it's not that powerful, but does have a reasonable usage in such case.

there're three builtin functions:

- 'view(..)': print every argument using fmt.Println
- 'read(int)': will read n bytes from underlying reader and return an array.
- 'fill(int)': will write n bytes to underlying writer.

and two variable:

- 'root': point to the top struct.
- 'current': point to the current struct.
- 'k': tell the outer slice is reading at which count. if nested, the closest one.

```go
Int16 int16 `rdm:"read(4)"` // read pre
Int16 int16 `rdn:"read(4)"` // read post
Int16 int16 `wtm:"read(4)"` // write pre
Int16 int16 `wtn:"read(4)"` // write post

Int16 []struct {
	Int16 `rdm:"root.Int16 == current.Int16, k"` // k will tell which n-th struct it is in
}
```

btw, stack and external variable map just have a size of 256. It's not to be modified by design.

# slice

in bstruct, special tags apply to slice.

for convenience, slice has three **reading** modes:

- modelen: mark tag 'length' as a program. prog should return the slice length before reading. zero length skip. and if elem is of basic type, it speeds up by a hack, but took a double space. array internally is implemented as a modelen slice.
- modesize: mark tag 'size' as a program. prog should return the space in bytes before readning. zero size skip. and if elem is of basic type, it becomes modelen. or it's modeeof with a limitedreader.
- modeeof: it will read until EOF, EOF is not an error.

as modesize, modeeof is not able to prealloc the slice, growing slice will leave useless buffer there, which may occupy times of space than the original data size.

there're two global variables can be tuned:

- SliceAccelThreshold: modelen does not pick the hack method until it has a larger length.
- SliceInitLen: modeof, the initial guess length.

```go
Int16 []int16
Int16 []int16 `length:"root.Int8"`
Int16 []int16 `size:"root.Int8"` // fallbacks to modelen
Int16 []struct{
	A [4]byte
} `size:"root.Int8"` // but not this one, so more space is taken
```

in these programs, however, `current` is not pointed to the parent struct of the current field, but the elem itself.

# string

as string is immutable, it does not implement slice mode. string is seen as a separate type, and is skipped automatically when read. the only way to change its value is assign.

# other thins

- endian: defaults to host endianess.
- elem of slice: because struct tag is only applied to field, so elem of slice is not seen as a field. tag therefore, is applied to slice itself rather than elem.
- nested slice/string: tag is applied to slice itself, so it's impossible to specify tag for the second dimesion slice. so nested slice/string is not allowed. but you can wrap it as a struct, then it's ok.
