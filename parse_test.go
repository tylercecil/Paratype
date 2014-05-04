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

func TestFail1(t *testing.T) {
	tclist, tlist, flist, err := paraparse.Setup("typeclass Num inherits Zin\ntypeclass Yin\ntype z implements Yin\nfunc foo(d, A) iNT\n=d\n")
	if err != nil {
		t.Error(err)
		return
	}
	PrintData(tclist, tlist, flist)
}

