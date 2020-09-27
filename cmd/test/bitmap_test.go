package test

import (
	"fmt"
	"testing"

	"github.com/RoaringBitmap/gocroaring"
)

func TestBitmap(t *testing.T) {
	b := gocroaring.New(1, 2, 3)
	r, err := b.Select(0)
	if nil != err {
		fmt.Println("err: ", err)
	} else {
		fmt.Println("ok: ", r)

	}
	c := gocroaring.New()
	fmt.Println(c.IsEmpty())
	fmt.Println(c.GetCardinality())
}

func TestBitmapOrAnd(t *testing.T) {
	b := gocroaring.New()

	b1 := gocroaring.New(1, 2, 3, 4, 5)
	b.Or(b1)
	b2 := gocroaring.New(2, 3)
	b.And(b2)
	b3 := gocroaring.New(3)
	b.And(b3)
	r, err := b.Select(0)

	fmt.Println("yes: ", r, "err:", err)
}
