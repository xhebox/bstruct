# bstruct

zero-copy, pre-generated binary encoder/decoder in golang with customization.

## performance


```
type Struct1 struct {
	A	bool
	B	[]bool
	C	Struct2
	D	string
	G	[]string
	E	Slice1
	F	bool
}
type Struct2 struct {
	A bool
}
type Slice1 []struct {
	E string
}
```

```
BenchmarkEncode-8      	17996954	        67.08 ns/op	      48 B/op	       2 allocs/op
BenchmarkMarshal-8     	 3224440	       372.8 ns/op	      96 B/op	       1 allocs/op
BenchmarkDecode-8      	22132478	        53.78 ns/op	      48 B/op	       1 allocs/op
BenchmarkUnmarshal-8   	  791223	      1493 ns/op	     272 B/op	       7 allocs/op
```
