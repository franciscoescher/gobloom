package main

import (
	"fmt"

	"github.com/franciscoescher/gobloom"
)

func main() {
	bf, _ := gobloom.NewScalable(gobloom.ParamsScalable{
		InitialSize:         1000,
		FalsePositiveRate:   0.01,
		FalsePositiveGrowth: 2,
	})
	bf.Add([]byte("foo"))
	bf.Add([]byte("bar"))
	bf.Add([]byte("baz"))

	fmt.Println(bf.Test([]byte("foo"))) // true
	fmt.Println(bf.Test([]byte("bar"))) // true
	fmt.Println(bf.Test([]byte("baz"))) // true
	fmt.Println(bf.Test([]byte("qux"))) // false
}
