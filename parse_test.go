package main

import (
	"Paratype/paraparse"
	"fmt"
	"testing"
)

func TestParsing(t *testing.T) {
	tclist, tlist, flist, err := paraparse.Setup("typeclass Num inherits Zin\ntypeclass Zin\ntype z implements Zin\nfunc foo(d, A) iNT\n=x\n")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("TCLIST: %+v\n", tclist)
	fmt.Printf("TLIST: %+v\n", tlist)
	fmt.Printf("FLIST: %+v\n", flist)
}