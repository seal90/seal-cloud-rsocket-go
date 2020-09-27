package test

import (
	"container/list"
	"fmt"
	"testing"
)

func TestList(t *testing.T) {
	l := list.New()
	l.PushBack(1)
	l.PushBack(2)
	fmt.Println(l.Back().Value)
}
