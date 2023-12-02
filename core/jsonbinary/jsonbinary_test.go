package jsonbinary

import (
	"core/binary"
	"core/checks"
	"core/variant"
	"fmt"
	"testing"

	pool "github.com/libp2p/go-buffer-pool"
)

type A struct {
	A0 string
	A1 binary.Binary
	A2 binary.Binary
}

func TestToTyped(t *testing.T) {
	a := A{
		A0: "Hello",
		A1: binary.FromBytes([]byte{'A', 'B', 'C'}),
		A2: binary.FromBytes([]byte{'D', 'E'}),
	}

	b, err := Serialize(&a, nil)

	var aa A
	err = Deserialize(b, &aa)
	if err != nil {
		panic(err)
	}
	checks.Equal(t, "Hello", aa.A0, "String field")
	checks.Equal(t, []byte{'A', 'B', 'C'}, aa.A1.Bytes(), "First binary field")
	checks.Equal(t, []byte{'D', 'E'}, aa.A2.Bytes(), "Subsequent binary field")
}

func TestToAny(t *testing.T) {
	a := A{
		A0: "Hello",
		A1: binary.FromBytes([]byte{'A', 'B', 'C'}),
		A2: binary.FromBytes([]byte{'D', 'E'}),
	}

	b, err := Serialize(&a, nil)

	fmt.Printf("B %v\n", string(b))

	var aa any
	err = Deserialize(b, &aa)
	if err != nil {
		panic(err)
	}
	var aaa A
	variant.Assign(aa, &aaa)
	checks.Equal(t, "Hello", aaa.A0, "String field")
	checks.Equal(t, []byte{'A', 'B', 'C'}, aaa.A1.Bytes(), "First binary field")
	checks.Equal(t, []byte{'D', 'E'}, aaa.A2.Bytes(), "Subsequent binary field")
}

func BenchmarkJsonBinary(bench *testing.B) {
	sizes := []int{-1, 0, 20, 100, 1000, 10000, 100000, 1000000}
	p := new(pool.BufferPool)
	p2 := BytesPool(p)
	pools := []BytesPool{nil, p2}

	for _, pl := range pools {
		poolStr := "no"
		if pl != nil {
			poolStr = "yes"
		}
		for _, size := range sizes {
			sizeStr := "none"

			a := A{
				A0: "Hello",
			}
			if size >= 0 {
				a.A1 = binary.FromBytes(make([]byte, size))
				sizeStr = fmt.Sprintf("%v", size)
			}

			bench.Run(fmt.Sprintf("Pool=%v,Length=%v", poolStr, sizeStr), func(bench *testing.B) {
				for i := 0; i < bench.N; i++ {
					b, err := Serialize(&a, pl)
					if err != nil {
						panic(err)
					}

					var aa any
					err = Deserialize(b, &aa)
					if err != nil {
						panic(err)
					}
					if pl != nil {
						pl.Put(b)
					}
				}
			})

		}
	}
}
