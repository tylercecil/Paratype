package main

import (
	"Paratype/paraparse"
	//"Paratype/context"
	"Paratype"
	//"fmt"
	"testing"
)


// Used to run a test and check the error. The data is then printed.
func RunTest(code string, t *testing.T) {
	flist, err := paraparse.Setup(code)
	if err != nil {
		t.Error(err)
		return
	}

	/*fmt.Printf("FLIST: %+v\n", flist)
	for _, f := range flist {
		fmt.Printf("%+v\n", f)
		context.PrintAll(f)
	}*/
	main.RunThem(1, flist)
}

// PASS
// typeclass Num inherits Zin
// typeclass Zin
// type z implements Zin
// func foo(d, A) iNT
// 		=x
func Test1(t *testing.T) {
	RunTest("typeclass Num inherits Zin\ntypeclass Zin\ntype z implements Zin\nfunc foo(d, A) iNT\n=x\n", t)
}

// PASS
// typeclass Num
// typeclass Zun
// type y implements Zun, Num
// type z implements Num
// func foo constrain A <Num, Zun> (d, A, y) iNT throws bigError, gError
// 		=x
func Test2(t *testing.T) {
	RunTest("typeclass Num\ntypeclass Zun\ntype y implements Zun, Num\ntype z implements Num\nfunc foo constrain A <Num, Zun> (d, A, y) iNT throws bigError, gError\n=x\n", t)
}

// PASS
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
	RunTest("typeclass Num\ntypeclass Zun\ntype y implements Zun, Num\ntype z implements Num\nfunc foo constrain A <Num, Zun> (d, A, y) iNt throws bigError, gError\n=x\nfunc goo(x, y) float throws veryBigError, someError, moreError\n=x\n", t)
}

// PASS
// typeclass Num
// type y implements Num
// func foo constrain A <Num> (y) int
// 		= y
func TestSimpleConstraint(t *testing.T) {
	RunTest("typeclass Num\ntype y implements Num\nfunc foo constrain A <Num> (y) int\n=y\n", t)
}

// PASS
// typeclass Num
// func foo(d, A) int
// 		=bar(A)
// func bar(int) B
// 		=int
func TestComposedFuncs(t *testing.T) {
	RunTest("typeclass Num\ntype int\ntype d\nfunc foo(d, A) int\n=bar(A)\nfunc bar(int) B\n=int\n", t)
}

// PASS
// func foo(d, A) int
// 		=bar(baz(ban(A)))
// func bar(int) B
// 		=int
// func baz(int) B
//		=int
// func ban(int) B
//		=baq()
// func baq() B
//		=B
func TestParents(t *testing.T) {
	RunTest("type d\ntype int\nfunc foo(d, A) int\n=bar(baz(ban(A)))\nfunc bar(int) B\n=int\nfunc baz(int) B\n=int\nfunc ban(int) B\n=baq()\nfunc baq() B\n=B\n", t)
}

func TestReallySimple(t *testing.T) {
	RunTest("func foo(A) A\n=A\n", t)
}

func TestFavorite(t *testing.T) {
	RunTest("typeclass Zat inherits Num\ntypeclass Num\ntype float implements Num\ntype int implements Num\nfunc f() float\n=g(int)\nfunc q() int\n=g(float)\nfunc g constrain A<Zat> (A) B\n=h(A)\nfunc h(A) B\n=B\n", t)

}

// PASS
// Tests correct creation of the children map.
// func foo(A) int
// 		=bar(baz(A), float)
// func bar(int, float) int
// 		=int
// func baz(int) A
// 		=int
func TestChildren(t *testing.T) {
	RunTest("func foo(A) int\n=bar(baz(A), float)\nfunc bar(int, float) int\n=int\nfunc baz(int) int\n=int\n", t)
}

// FAIL
// This should fail. Check if a non-existent function fails correctly.
// func foo(A) int
// 		=bar(bal(A), float)
// func bar(int, float) int
// 		=int
// func baz(int) int
// 		=int
func TestNonExistentChild(t *testing.T) {
	RunTest("func foo(A) int\n=bar(bal(A), float)\nfunc bar(int, float) int\n=int\nfunc baz(int) int\n=int\n", t)
}

// PASS
// Test the situation where a function has one child function.
// func foo(A) int
// 		=bar(A)
// func bar(A) int
// 		=int
func TestOnlyChild(t *testing.T) {
	RunTest("func foo(A) int\n=bar(A)\nfunc bar(A) int\n=int\n", t)
}

// Should fail because Zin does not exist.
// typeclass Num inherits Zin
// typeclass Yin
// type z implements Yin
// func foo(d, A) iNT
// 		=d
func TestFail1(t *testing.T) {
	RunTest("typeclass Num inherits Zin\ntypeclass Yin\ntype z implements Yin\nfunc foo(d, A) iNT\n=d\n", t)
}

