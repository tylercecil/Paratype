package context

import (
	"bytes"
	"strconv"
)

func ConvertPath(f []*Function) string {
	var buf bytes.Buffer

	for _, fun := range f {
		buf.WriteString(strconv.Itoa(fun.id))
		buf.WriteString("-")
	}
	s := buf.String()
	s = s[:len(s)-1] // remove last character
	return s
}

// updates typevar v in g to be typevar w in f
func (g *Function) updateTypevar(v *TypeVariable, f *Function, w *TypeVariable) {
	//
	if f.typeMap[w] != nil && g.typeMap[v] != nil && f.typeMap[w] != g.typeMap[v] {
		// TYPE ERROR
	}

	// find explicit type if it exists (nil otherwise)
	if g.typeMap[v] != nil {
		g.typeMap[w] = g.typeMap[v]
	} else {
		g.typeMap[w] = f.typeMap[w]
	}

	// intersection of w.constraints and v.constraints
	if w.constraints[nil] { // w allows any
		w.constraints = v.constraints
	} else if v.constraints[nil] == false {
		for typeclass := range w.constraints {
			if v.constraints[typeclass] == false {
				w.constraints[typeclass] = false
			}
		}
	}

	// is new explicit type adhering to type constraints?
	if len(w.constraints) == 0 {
		// TYPE ERROR
	} else {
		impl := false
		for typeclass := range w.constraints {
			if g.typeMap[w].implements[typeclass] {
				impl = true
			}
		}
		if impl == false {
			// TYPE ERROR
		}
	}

	f.typeVarMap[v] = w
}


func (f *Function) Update(g *Function) {
	// lock both f and g

	var pf = ConvertPath([]*Function{f})
	var pgf = ConvertPath([]*Function{g, f})
	var pg = ConvertPath([]*Function{g})

	// f is child of g
	if g.children[&f.Context] {
		for funcarg, typevar := range f.atlas[pf] {
			g.updateTypevar(g.atlas[pgf][funcarg], f, typevar)
		}

		f.parents[&g.Context] = true
	}

	for funcarg, typevar := range g.atlas[pg] {
		if f.typeVarMap[typevar] != nil {
			g.updateTypevar(typevar, f, f.atlas[pf][funcarg])
		}
	}

	// E_g = E_g union E_f
	for errorType := range f.errors {
		g.errors[errorType] = true
	}
}
