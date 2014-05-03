package main

import (
	"testing"
	"fmt"
	"Paratype/context"
)

var funcCounter int = 0

func MakeTestTypes() (num context.TypeClass, mat context.TypeClass, in context.Type, fl context.Type) {
	s := map[*context.TypeClass]bool{nil : true}
	num = context.TypeClass{"Num", s};
	p := map[*context.TypeClass]bool{nil : true}
	mat	= context.TypeClass{"Matrix", p};

	in = context.Type{"Int", map[*context.TypeClass]bool{&num : true, nil : true}};
	fl = context.Type{"Float", map[*context.TypeClass]bool{&num : true, nil : true}};
	return
}

func MakeFunction(name string, numArgs int) context.Function {
	g := context.Function{
		Name : name,
		Id: funcCounter,
		NumArgs : numArgs,
		Context : context.Context{
			Atlas: map[string]map[int]*context.TypeVariable{},
			TypeMap: map[*context.TypeVariable]*context.Type{},
			TypeVarMap: map[*context.TypeVariable]*context.TypeVariable{},
			Errors : map[*context.Type]bool{},
			Parents : map[*context.Context]bool{},
			Children : map[*context.Context]bool{},
		},
	}
	funcCounter++
	return g
}

func MakeTypeVar(name string, res bool) context.TypeVariable {
	return context.TypeVariable{
		Constraints : map[*context.TypeClass]bool{nil : true},
		Resolved	: res,
		Name		: name,
	}
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

func TestMergeDown(t *testing.T) {
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
	g.TypeMap[&G0] = nil
	g.TypeMap[&G1] = nil
	g.TypeMap[&G2] = nil

	f := MakeFunction("f", 1)
	f.TypeMap[&F0] = &in
	f.TypeMap[&F1] = &fl
	f.TypeMap[&F2] = &in

	pf := context.ConvertPath([]*context.Function{&f})
	pfg := context.ConvertPath([]*context.Function{&f, &g})
	pg := context.ConvertPath([]*context.Function{&g})

	f.Atlas[pf] = map[int]*context.TypeVariable{0 : &F0}
	f.Atlas[pfg] = map[int]*context.TypeVariable{0 : &F0, 1 : &F1, 2 : &F2}
	g.Atlas[pg] = map[int]*context.TypeVariable{0 : &G0, 1 : &G1, 2 : &G2}
	f.Children[&g.Context] = true

	fmt.Printf("\nTypemap of f\n")
	PrintTypeMap(&f)
	fmt.Printf("\nAtlas of f\n")
	PrintAtlas(&f)
	fmt.Printf("\nTypemap of g\n")
	PrintTypeMap(&g)
	fmt.Printf("\nAtlas of g\n")
	PrintAtlas(&g)

	g.Update(&f)

	fmt.Printf("\nTypemap of f\n")
	PrintTypeMap(&f)
	fmt.Printf("\nAtlas of f\n")
	PrintAtlas(&f)
}

func TestMergeUp(t *testing.T) {
	// func f constraint T<Num> (T R) S
	//  = g(T R)
	// func g(int float) int
	//  = int
	// f : F_0 F_1 F_2
	// f \circ g : F_0 F_1 F_2
	// g : G_0 G_1 G_2
	DownExample(0) // explicit type conflict
	DownExample(1) // typeclass conflict
	DownExample(2) // explicit type not in merged typeclass
	DownExample(3) // no error
}


func DownExample(errcode int) {
	num, mat, in, fl := MakeTestTypes()

	var F0 context.TypeVariable
	var G0 context.TypeVariable

	if errcode == 2 || errcode == 1 {
		F0 = context.TypeVariable{
			Constraints : map[*context.TypeClass]bool{&mat : true},
			Resolved	: false,
			Name		: "F_0",
		}
	} else {
		F0 = MakeTypeVar("F_0", false)
	}

	F1 := MakeTypeVar("F_1", false)
	F2 := MakeTypeVar("F_2", false)

	if errcode == 1 {
		G0 = context.TypeVariable{
			Constraints : map[*context.TypeClass]bool{&num : true},
			Resolved	: true,
			Name		: "G_0",
		}
	} else {
		G0 = MakeTypeVar("G_0", true)
	}

	G1 := MakeTypeVar("G_1", true)
	G2 := MakeTypeVar("G_2", true)

	g := MakeFunction("g", 3)
	g.TypeMap[&G0] = &in
	g.TypeMap[&G1] = &fl
	g.TypeMap[&G2] = &in

	f := MakeFunction("f", 3)
	if errcode == 0 {
		f.TypeMap[&F0] = &fl
	} else {
		f.TypeMap[&F0] = nil
	}
	f.TypeMap[&F1] = nil
	f.TypeMap[&F2] = nil

	pf := context.ConvertPath([]*context.Function{&f})
	pfg := context.ConvertPath([]*context.Function{&f, &g})
	pg := context.ConvertPath([]*context.Function{&g})

	f.Atlas[pf] = map[int]*context.TypeVariable{0 : &F0, 1 : &F1, 2 : &F2}
	f.Atlas[pfg] = map[int]*context.TypeVariable{0 : &F0, 1 : &F1, 2 : &F2}
	g.Atlas[pg] = map[int]*context.TypeVariable{0 : &G0, 1 : &G1, 2 : &G2}
	f.Children[&g.Context] = true

	/*fmt.Printf("\nTypemap of f\n")
	PrintTypeMap(&f)
	fmt.Printf("\nAtlas of f\n")
	PrintAtlas(&f)
	fmt.Printf("\nTypemap of g\n")
	PrintTypeMap(&g)
	fmt.Printf("\nAtlas of g\n")
	PrintAtlas(&g)*/

	g.Update(&f)

	/*fmt.Printf("\nTypemap of f\n")
	PrintTypeMap(&f)
	fmt.Printf("\nAtlas of f\n")
	PrintAtlas(&f)
	fmt.Printf("\nTypemap of g\n")
	PrintTypeMap(&g)
	fmt.Printf("\nAtlas of g\n")
	PrintAtlas(&g)*/
}
