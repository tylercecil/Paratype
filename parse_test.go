package main

import (
	"Paratype/paraparse"
	"Paratype/context"
	"fmt"
	"testing"
)

func PrintData(tclist []context.TypeClass, tlist []context.Type, flist []context.Function) {
	fmt.Printf("TCLIST: %+v\n", tclist)
	fmt.Printf("TLIST: %+v\n", tlist)
	fmt.Printf("FLIST: %+v\n", flist)
}

func Test1(t *testing.T) {
	tclist, tlist, flist, err := paraparse.Setup("typeclass Num inherits Zin\ntypeclass Zin\ntype z implements Zin\nfunc foo(d, A) iNT\n=x\n")
	if err != nil {
		t.Error(err)
		return
	}
	PrintData(tclist, tlist, flist)
}

func Test2(t *testing.T) {
	tclist, tlist, flist, err := paraparse.Setup("typeclass Num\ntypeclass Zun\ntype y implements Zun, Num\ntype z implements Num\nfunc foo constrain A <Num, Zun> (d, A, y) iNT throws bigError, gError\n=x\n")
	if err != nil {
		t.Error(err)
		return
	}
	PrintData(tclist, tlist, flist)
}

// FOO: 4 Parameters
// GOO: 3 Parameters
// typeclass Num
// typeclass Zun
// type y implements Zun, Num
// type z implements Num
// func foo constrain A <Num, Zun> (d, A, y) iNt throws bigError, gError
// 		=x
// func goo(x, y) float throws veryBigError, someError, moreError
// 		=x
func TestMultFuncs(t *testing.T) {
	tclist, tlist, flist, err := paraparse.Setup("typeclass Num\ntypeclass Zun\ntype y implements Zun, Num\ntype z implements Num\nfunc foo constrain A <Num, Zun> (d, A, y) iNt throws bigError, gError\n=x\nfunc goo(x, y) float throws veryBigError, someError, moreError\n=x\n")
	if err != nil {
		t.Error(err)
		return
	}
	PrintData(tclist, tlist, flist)
}

// Should fail because Zin does not exist.
// typeclass Num inherits Zin
// typeclass Yin
// type z implements Yin
// func foo(d, A) iNT
// 		=d
func TestFail1(t *testing.T) {
	tclist, tlist, flist, err := paraparse.Setup("typeclass Num inherits Zin\ntypeclass Yin\ntype z implements Yin\nfunc foo(d, A) iNT\n=d\n")
	if err != nil {
		t.Error(err)
		return
	}
	PrintData(tclist, tlist, flist)
}

