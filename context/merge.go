package context

import (
	"bytes"
	"strconv"
)

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
func (g *Function) updateTypevar(v *TypeVariable, f *Function, w *TypeVariable) {
	//
	if f.TypeMap[w] != nil && g.TypeMap[v] != nil && f.TypeMap[w] != g.TypeMap[v] {
		// TYPE ERROR
	}

	// find explicit type if it exists (nil otherwise)
	if g.TypeMap[v] != nil {
		g.TypeMap[w] = g.TypeMap[v]
	} else {
		g.TypeMap[w] = f.TypeMap[w]
	}

	// intersection of w.Constraints and v.Constraints
	if w.Constraints[nil] { // w allows any
		w.Constraints = v.Constraints
	} else if v.Constraints[nil] == false {
		for typeclass := range w.Constraints {
			if v.Constraints[typeclass] == false {
				w.Constraints[typeclass] = false
			}
		}
	}

	// is new explicit type adhering to type Constraints?
	if len(w.Constraints) == 0 {
		// TYPE ERROR
	} else {
		impl := false
		for typeclass := range w.Constraints {
			if g.TypeMap[w].Implements[typeclass] {
				impl = true
			}
		}
		if impl == false {
			// TYPE ERROR
		}
	}

	f.TypeVarMap[v] = w
}


func (f *Function) Update(g *Function) {
	// lock both f and g

	var pf = ConvertPath([]*Function{f})
	var pgf = ConvertPath([]*Function{g, f})
	var pg = ConvertPath([]*Function{g})

	// f is child of g
	if g.Children[&f.Context] {
		for funcarg, typevar := range f.Atlas[pf] {
			g.updateTypevar(g.Atlas[pgf][funcarg], f, typevar)
		}

		f.Parents[&g.Context] = true
	}

	for funcarg, typevar := range g.Atlas[pg] {
		if f.TypeVarMap[typevar] != nil {
			g.updateTypevar(typevar, f, f.Atlas[pf][funcarg])
		}
	}

	// E_g = E_g union E_f
	for errorType := range f.Errors {
		g.Errors[errorType] = true
	}
}
