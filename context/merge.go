package context


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
		impl = false
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

	// f is child of g
	/*if g.children[f] {
		for funcarg, typevar := range f.atlas[f] {
			g.updateTypevar(g.atlas[g of f][funcarg], f, typevar)
		}

		f.parents[g] = true
	}

	for funcarg, typevar := g.atlas[g] {
		if f.typeVarMap[typevar] != nil {
			g.updateTypevar(typevar, f, f.atlas[f][funcarg])
		}
	}*/

	// E_g = E_g union E_f
	for errorType := range f.errors {
		g.errors[errorType] = true
	}
}
