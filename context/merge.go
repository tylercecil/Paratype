package context

import (
	"bytes"
	"strconv"
	"fmt"
)

var errors = [...]string{
	"Explicit type conflict",
	"Type class conflict",
	"Explicit type not of merged type class",
};

func PrintError(errcode int, f *Function, g *Function) {
	fmt.Printf("\n===TYPE ERROR %v===\n%v in %v when merged with %v\n\n", errcode, errors[errcode], f.Name, g.Name)
}

func ConvertPath(f []*Function) string {
	var buf bytes.Buffer

	for _, fun := range f {
		buf.WriteString(strconv.Itoa(fun.Id))
		buf.WriteString("-")
	}
	s := buf.String()
	s = s[:len(s)-1] // remove last character
	return s
}

// updates typevar v in g to be typevar w in f
func (g *Function) updateTypevar(path string, funcarg int, f *Function, w *TypeVariable) {
	v := g.Atlas[path][funcarg]
	if f.TypeMap[w] != nil && g.TypeMap[v] != nil && f.TypeMap[w] != g.TypeMap[v] {
		// explicit types do not match
		PrintError(0, g, f)
	}

	// find explicit type if it exists (nil otherwise)
	if g.TypeMap[v] != nil {
		g.TypeMap[w] = g.TypeMap[v]
	} else {
		g.TypeMap[w] = f.TypeMap[w]
	}

	if g.TypeMap[w] != nil {
		w.Resolved = true
	}

	// intersection of w.Constraints and v.Constraints
	if w.Constraints[nil] { // w allows any
		w.Constraints = v.Constraints
	} else if v.Constraints[nil] == false {
		for typeclass := range w.Constraints {
			if v.Constraints[typeclass] == false {
				delete(w.Constraints, typeclass)
			}
		}
	}

	// is new explicit type adhering to type Constraints?
	if len(w.Constraints) == 0 {
		// merging typeclasses brought us no typeclasses.
		PrintError(1, g, f)
	} else if w.Constraints[nil] == false {
		impl := false
		for typeclass := range w.Constraints {
			if g.TypeMap[w].Implements[typeclass] {
				impl = true
			}
		}

		// the explicit type does not implemented any of the allowed merged
		// TypeClasses
		if impl == false {
			PrintError(2, g, f)
		}
	}

	f.TypeVarMap[v] = w
	g.Atlas[path][funcarg] = w
}


func (f *Function) Update(g *Function) {
	// lock both f and g

	var pf = ConvertPath([]*Function{f})
	var pgf = ConvertPath([]*Function{g, f})
	var pg = ConvertPath([]*Function{g})

	// f is child of g
	if g.Children[&f.Context] {
		for funcarg, typevar := range f.Atlas[pf] {
			g.updateTypevar(pgf, funcarg, f, typevar)
		}

		f.Parents[&g.Context] = true
	}

	for funcarg, typevar := range g.Atlas[pg] {
		if f.TypeVarMap[typevar] != nil {
			g.updateTypevar(pg, funcarg, f, f.Atlas[pf][funcarg])
		}
	}

	// E_g = E_g union E_f
	for errorType := range f.Errors {
		g.Errors[errorType] = true
	}
}
