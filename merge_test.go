package main

import (
	"testing"
	"fmt"
	"Paratype/context"
	"Paratype"
)

var funcCounter int = 0

func PrintAll(f *context.Function) {
	fmt.Printf("\nTypemap of %v\n", f.Name)
	PrintTypeMap(f)
	fmt.Printf("\nAtlas of %v\n", f.Name)
	PrintAtlas(f)
}

func MakeTestTypes() (num *context.TypeClass, mat *context.TypeClass,
	in *context.Type, fl *context.Type) {
	s := make(map[*context.TypeClass]bool)
	s[nil] = true
	num = new(context.TypeClass)
	num.Name = "Num"
	num.Inherits = s
	p := make(map[*context.TypeClass]bool)
	p[nil] = true
	mat	= new(context.TypeClass)
	mat.Name = "Matrix"
	mat.Inherits = p;

	ma := make(map[*context.TypeClass]bool)
	ma[num] = true
	ma[nil] = true

	in = new(context.Type)
	in.Name = "Int"
	in.Implements = ma

	fl = new(context.Type)
	fl.Name = "Float"
	fl.Implements = ma
	return
}

func MakeFunction(name string, numArgs int) *context.Function {
	g := new(context.Function)
	g.Name = name
	g.Id = funcCounter
	g.NumArgs = numArgs
	g.Atlas = make(map[string]map[int]*context.TypeVariable)
	g.TypeMap = make(map[*context.TypeVariable]*context.Type)
	g.TypeVarMap = make(map[*context.TypeVariable]*context.TypeVariable)
	g.Errors = make(map[*context.Type]bool)
	g.Parents = make(map[*context.Function]bool)
	g.Children = make(map[int]map[*context.Function]bool)
	funcCounter++
	return g
}

func MakeTypeVar(name string, res bool) *context.TypeVariable {
	s := new(context.TypeVariable)
	s.Constraints = make(map[*context.TypeClass]bool)
	s.Constraints[nil] = true
	s.Resolved = res
	s.Name = name
	return s
}

func PrintTypeMap(g *context.Function) {
	for tv, t := range g.TypeMap {
		fmt.Printf("%+v : %+v\n", tv, t)
	}
}

func PrintAtlas(g *context.Function) {
	for path, tuple := range g.Atlas {
		fmt.Printf("%+v\n", path)
		for _, tv := range tuple {
			fmt.Printf("%+v\n", tv)
		}
	}
}

// Test of: f calls g, f has explicit types
func TestDown(t *testing.T) {
	// func f() Int
	//  = g(Int Float)
	// func g(T R) S
	//  = S
	// f : F_0
	// f \circ g : F_0 F_1 F_2
	// g : G_0 G_1 G_2
	_, _, in, fl := MakeTestTypes()
	F0 := MakeTypeVar("F_0", true)
	F1 := MakeTypeVar("F_1", true)
	F2 := MakeTypeVar("F_2", true)
	G0 := MakeTypeVar("G_0", false)
	G1 := MakeTypeVar("G_1", false)
	G2 := MakeTypeVar("G_2", false)

	g := MakeFunction("g", 3)
	g.TypeMap[G0] = nil
	g.TypeMap[G1] = nil
	g.TypeMap[G2] = nil

	f := MakeFunction("f", 1)
	f.TypeMap[F0] = in
	f.TypeMap[F1] = fl
	f.TypeMap[F2] = in

	pf := context.FunctionsToPath(f)
	pfg := context.FunctionsToPath(f, g)
	pg := context.FunctionsToPath(g)

	f.Atlas[pf] = map[int]*context.TypeVariable{0 : F0}
	f.Atlas[pfg] = map[int]*context.TypeVariable{0 : F0, 1 : F1, 2 : F2}
	g.Atlas[pg] = map[int]*context.TypeVariable{0 : G0, 1 : G1, 2 : G2}
	f.Children[0] = make(map[*context.Function]bool)
	f.Children[0][g] = true

	//PrintAll(f)
	//PrintAll(g)

	//g.Update(f)
	main.RunThings(f, g)

	//PrintAll(f)
	//PrintAll(g)

	fmt.Printf("\n===implementations===\n\n")
	f.Finish()
	g.Finish()
	fmt.Printf("\n")

}

// f calls g, g has explicit types
func TestUp0(t *testing.T) {
	// func f constraint T<Num> (T R) S
	//  = g(T R)
	// func g(int float) int
	//  = int
	// f : F_0 F_1 F_2
	// f \circ g : F_0 F_1 F_2
	// g : G_0 G_1 G_2
	DownExample(0, t) // explicit type conflict (F_0 fl, G_0 in)
}

func TestUp1(t *testing.T) {
	DownExample(1, t) // typeclass conflict
}

func TestUp2(t *testing.T) {
	DownExample(2, t) // explicit type not in merged typeclass (in not mat)
}

func TestUp3(t *testing.T) {
	DownExample(3, t) // no error
}

func DownExample(errcode int, t * testing.T) {
	num, mat, in, fl := MakeTestTypes()

	F0 := MakeTypeVar("F_0", false)
	G0 := MakeTypeVar("G_0", true)

	delete(F0.Constraints, nil)
	if errcode == 2 || errcode == 1 {
		F0.Constraints[mat] = true
	} else {
		F0.Constraints[num] = true
	}

	F1 := MakeTypeVar("F_1", false)
	F2 := MakeTypeVar("F_2", false)

	if errcode == 1 {
		delete(G0.Constraints, nil)
		G0.Constraints[num] = true
	}

	G1 := MakeTypeVar("G_1", true)
	G2 := MakeTypeVar("G_2", true)

	g := MakeFunction("g", 3)
	g.TypeMap[G0] = in
	g.TypeMap[G1] = fl
	g.TypeMap[G2] = in

	f := MakeFunction("f", 3)
	if errcode == 0 {
		f.TypeMap[F0] = fl
	} else {
		f.TypeMap[F0] = nil
	}
	f.TypeMap[F1] = nil
	f.TypeMap[F2] = nil

	pf := context.FunctionsToPath(f)
	pfg := context.FunctionsToPath(f, g)
	pg := context.FunctionsToPath(g)

	f.Atlas[pf] = map[int]*context.TypeVariable{0 : F0, 1 : F1, 2 : F2}
	f.Atlas[pfg] = map[int]*context.TypeVariable{0 : F0, 1 : F1, 2 : F2}
	g.Atlas[pg] = map[int]*context.TypeVariable{0 : G0, 1 : G1, 2 : G2}
	f.Children[0] = make(map[*context.Function]bool)
	f.Children[0][g] = true

	//PrintAll(f)
	//PrintAll(g)

	//g.Update(f)
	main.RunThings(f, g)


	//PrintAll(f)
	//PrintAll(g)

	fmt.Printf("\n===implementations===\n\n")
	f.Finish()
	g.Finish()
	fmt.Printf("\n")
}


// g and h call f, mixed explicit types
func TestTwo(t *testing.T) {
	TwoExample(0, t)
}

func TwoExample(errcode int, t * testing.T) {
	_, _, in, fl := MakeTestTypes()

	// f(T) float
	// = float
	//
	// g(int) T
	// = f(int)
	//
	// h(int) T
	// = f(float)

	F0 := MakeTypeVar("F_0", true)

	/*delete(F0.Constraints, nil)
	if errcode == 2 || errcode == 1 {
		F0.Constraints[mat] = true
	} else {
		F0.Constraints[num] = true
	}*/

	F1 := MakeTypeVar("F_1", false)
	//F2 := MakeTypeVar("F_2", false)

	G0 := MakeTypeVar("G_0", false)
	/*if errcode == 1 {
		delete(G0.Constraints, nil)
		G0.Constraints[num] = true
	}*/

	G1 := MakeTypeVar("G_1", true)
	G2 := MakeTypeVar("G_2", true)

	H0 := MakeTypeVar("H_0", false)
	H1 := MakeTypeVar("H_1", true)
	H2 := MakeTypeVar("H_2", true)

	g := MakeFunction("g", 2)
	g.TypeMap[G0] = nil
	g.TypeMap[G1] = in
	g.TypeMap[G2] = in

	h := MakeFunction("h", 2)
	h.TypeMap[H0] = nil
	h.TypeMap[H1] = fl
	h.TypeMap[H2] = fl


	f := MakeFunction("f", 2)
	//if errcode == 0 {
	f.TypeMap[F0] = fl
	/*} else {
		f.TypeMap[F0] = nil
	}*/
	f.TypeMap[F1] = nil

	pf := context.FunctionsToPath(f)
	pgf := context.FunctionsToPath(g, f)
	pg := context.FunctionsToPath(g)
	phf := context.FunctionsToPath(h, f)
	ph := context.FunctionsToPath(h)

	f.Atlas[pf] = map[int]*context.TypeVariable{0 : F0, 1 : F1}
	g.Atlas[pgf] = map[int]*context.TypeVariable{0 : G0, 1 : G2}
	g.Atlas[pg] = map[int]*context.TypeVariable{0 : G0, 1 : G1}
	h.Atlas[phf] = map[int]*context.TypeVariable{0 : H0, 1 : H2}
	h.Atlas[ph] = map[int]*context.TypeVariable{0 : H0, 1 : H1}
	h.Children[0] = make(map[*context.Function]bool)
	g.Children[0] = make(map[*context.Function]bool)
	g.Children[0][f] = true
	h.Children[0][f] = true

	//PrintAll(f)
	//PrintAll(g)

	//f.Update(g)
	//f.Update(h)
	main.RunThings(f, g, h)

	/*PrintAll(f)
	PrintAll(g)
	PrintAll(h)*/

	fmt.Printf("\n===implementations===\n\n")
	f.Finish()
	g.Finish()
	h.Finish()
	fmt.Printf("\n")
}

func TestFlow(t *testing.T) {
	FlowExample(0, t)
}


func FlowExample(errcode int, t * testing.T) {
	_, _, in, fl := MakeTestTypes()

	// func f() float
	// = g(int)
	// 
	// func q() int
	// = g(float)
	//
	// func g(T) R
	// = h(T)
	//
	// func h(S) U
	// = U

	F0 := MakeTypeVar("F_0", true)
	F1 := MakeTypeVar("F_1", true)
	Q0 := MakeTypeVar("Q_0", true)
	Q1 := MakeTypeVar("Q_1", true)
	G0 := MakeTypeVar("G_0", false)
	G1 := MakeTypeVar("G_1", false)
	H0 := MakeTypeVar("H_0", false)
	H1 := MakeTypeVar("H_1", false)

	g := MakeFunction("g", 1)
	g.TypeMap[G0] = nil
	g.TypeMap[G1] = nil

	h := MakeFunction("h", 1)
	h.TypeMap[H0] = nil
	h.TypeMap[H1] = nil

	f := MakeFunction("f", 1)
	f.TypeMap[F0] = fl
	f.TypeMap[F1] = in

	q := MakeFunction("q", 1)
	q.TypeMap[Q0] = in
	q.TypeMap[Q1] = fl

	pf := context.FunctionsToPath(f)
	pfg := context.FunctionsToPath(f, g)
	pq := context.FunctionsToPath(q)
	pqg := context.FunctionsToPath(q, g)
	pg := context.FunctionsToPath(g)
	pgh := context.FunctionsToPath(g, h)
	ph := context.FunctionsToPath(h)

	f.Atlas[pf] = map[int]*context.TypeVariable{0 : F0}
	f.Atlas[pfg] = map[int]*context.TypeVariable{0 : F0, 1 : F1}
	q.Atlas[pq] = map[int]*context.TypeVariable{0 : Q0}
	q.Atlas[pqg] = map[int]*context.TypeVariable{0 : Q0, 1 : Q1}
	g.Atlas[pg] = map[int]*context.TypeVariable{0 : G0, 1 : G1}
	g.Atlas[pgh] = map[int]*context.TypeVariable{0 : G0, 1 : G1}
	h.Atlas[ph] = map[int]*context.TypeVariable{0 : H0, 1 : H1}
	f.Children[0] = make(map[*context.Function]bool)
	q.Children[0] = make(map[*context.Function]bool)
	g.Children[0] = make(map[*context.Function]bool)
	f.Children[0][g] = true
	q.Children[0][g] = true
	g.Children[0][h] = true

	main.RunThings(f, q, g, h)

	/*PrintAll(f)
	PrintAll(g)
	PrintAll(h)*/

	fmt.Printf("\n===implementations===\n\n")
	q.Finish()
	f.Finish()
	g.Finish()
	h.Finish()
	fmt.Printf("\n")
}
