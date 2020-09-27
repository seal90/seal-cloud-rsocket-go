package test

import (
	"fmt"
	"testing"
)

func TestOutRange(t *testing.T) {
	var a int8
	a = 127
	a++
	fmt.Println(a)
}
