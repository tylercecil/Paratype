package main

import (
	"testing"
	"fmt"
	"Paratype/context"
)

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
	s := map[*context.TypeClass]bool{nil : true}
	num := context.TypeClass{"Num", s};
	in := context.Type{"Int", map[*context.TypeClass]bool{&num : true, nil :
	true}};
	fl := context.Type{"Float", map[*context.TypeClass]bool{&num : true, nil :
true}};

	F0 := MakeTypeVar("F_0", true)
	F1 := MakeTypeVar("F_1", true)
	F2 := MakeTypeVar("F_2", true)
	G0 := MakeTypeVar("G_0", false)
	G1 := MakeTypeVar("G_1", false)
	G2 := MakeTypeVar("G_2", false)

	g := context.Function{
		Name : "g", Id: 0, NumArgs : 3, Context : context.Context{
			Atlas: map[string]map[int]*context.TypeVariable{},
			TypeMap: map[*context.TypeVariable]*context.Type{&G0 : nil, &G1 :
			nil, &G2 : nil},
			TypeVarMap: map[*context.TypeVariable]*context.TypeVariable{},
			Errors : map[*context.Type]bool{},
			Parents : map[*context.Context]bool{},
			Children : map[*context.Context]bool{},
		},
	}

	f := context.Function{
		Name : "f", Id: 1, NumArgs : 1, Context : context.Context{
			Atlas: map[string]map[int]*context.TypeVariable{},
			TypeMap: map[*context.TypeVariable]*context.Type{&F0 : &in, &F1 :
			&fl, &F2 : &in},
			TypeVarMap: map[*context.TypeVariable]*context.TypeVariable{},
			Errors : map[*context.Type]bool{},
			Parents : map[*context.Context]bool{},
			Children : map[*context.Context]bool{},
		},
	}

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
	// func f(T R) S
	//  = g(T R)
	// func g(int float) int
	//  = int
	// f : F_0
	// f \circ g : F_0 F_1 F_2
	// g : G_0 G_1 G_2
	s := map[*context.TypeClass]bool{nil : true}
	num := context.TypeClass{"Num", s};
	in := context.Type{"Int", map[*context.TypeClass]bool{&num : true, nil :
	true}};
	fl := context.Type{"Float", map[*context.TypeClass]bool{&num : true, nil :
true}};

	F0 := MakeTypeVar("F_0", false)
	F1 := MakeTypeVar("F_1", false)
	F2 := MakeTypeVar("F_2", false)
	G0 := MakeTypeVar("G_0", true)
	G1 := MakeTypeVar("G_1", true)
	G2 := MakeTypeVar("G_2", true)

	g := context.Function{
		Name : "g", Id: 0, NumArgs : 3, Context : context.Context{
			Atlas: map[string]map[int]*context.TypeVariable{},
			TypeMap: map[*context.TypeVariable]*context.Type{&G0 : &in, &G1 :
			&fl, &G2 : &in},
			TypeVarMap: map[*context.TypeVariable]*context.TypeVariable{},
			Errors : map[*context.Type]bool{},
			Parents : map[*context.Context]bool{},
			Children : map[*context.Context]bool{},
		},
	}

	f := context.Function{
		Name : "f", Id: 1, NumArgs : 1, Context : context.Context{
			Atlas: map[string]map[int]*context.TypeVariable{},
			TypeMap: map[*context.TypeVariable]*context.Type{&F0 : nil, &F1 :
			nil, &F2 : nil},
			TypeVarMap: map[*context.TypeVariable]*context.TypeVariable{},
			Errors : map[*context.Type]bool{},
			Parents : map[*context.Context]bool{},
			Children : map[*context.Context]bool{},
		},
	}

	pf := context.ConvertPath([]*context.Function{&f})
	pfg := context.ConvertPath([]*context.Function{&f, &g})
	pg := context.ConvertPath([]*context.Function{&g})
	f.Atlas[pf] = map[int]*context.TypeVariable{0 : &F0, 1 : &F1, 2 : &F2}
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
	fmt.Printf("\nTypemap of g\n")
	PrintTypeMap(&g)
	fmt.Printf("\nAtlas of g\n")
	PrintAtlas(&g)
}
