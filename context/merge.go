package context

import (
	"bytes"
	"strconv"
	"fmt"
	"errors"
)

var errorMsgs = [...]string{
	"Explicit type conflict",
	"Type class conflict",
	"Explicit type not of merged type class",
	"Missing implementation",
};

// helper function for printing type errors --- later, we need to account for
// stopping all threads in case of a type error
func PrintError(errcode int, f *Function, g *Function) string {
	switch {
	case errcode <= 2:
		return fmt.Sprintf("\n===TYPE ERROR %v===\n%v in %v when merged with %v\n\n",
			errcode, errorMsgs[errcode], f.Name, g.Name)
	case errcode == 3:
		return fmt.Sprintf("\n===TYPE ERROR %v===\n%v in %v\n\n",
			errcode, errorMsgs[errcode], f.Name)
	}
	return ""
}

// convert an array of function pointers to a path to be used as key for atlas
func ConvertPath(f ...interface{}) string {
	var buf bytes.Buffer

	for _, fun := range f {
		buf.WriteString(strconv.Itoa(fun.(*Function).Id))
		buf.WriteString("-")
	}
	s := buf.String()
	s = s[:len(s)-1] // remove last character
	return s
}

func IsChild(child *Function, parent *Function) int {
	for i, fmap := range parent.Children {
		if fmap[child] {
			return i
		}
	}
	return -1
}

// updates typevar v in func g to be typevar w in func f
// really, we are merging v and w to be w in g
func (g *Function) updateTypevar(path string, funcarg int, f *Function,
	w *TypeVariable) error {
	v := g.Atlas[path][funcarg]

	if f.TypeMap[w] != nil && g.TypeMap[v] != nil && f.TypeMap[w] != g.TypeMap[v] {
		// explicit types do not match
		return errors.New(PrintError(0, g, f))
	}

	// make typemap entry for w in g
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
	if w.Constraints[nil] {
		// w allows any typeclass
		w.Constraints = v.Constraints
	} else if v.Constraints[nil] == false {
		// neither w nor v allow any typeclass, so merge their actual
		// typeclassses
		for typeclass := range w.Constraints {
			if v.Constraints[typeclass] == false {
				delete(w.Constraints, typeclass)
			}
		}
	}

	if len(w.Constraints) == 0 {
		// merging typeclasses brought us no typeclasses -- not even nil (any)
		return errors.New(PrintError(1, g, f))
	} else if w.Constraints[nil] == false && g.TypeMap[w] != nil {
		// is new explicit type of w adhering to merged typeclass constraints?
		var impl *TypeClass = nil
		for typeclass := range w.Constraints {
			if g.TypeMap[w].Implements[typeclass] {
				impl = typeclass
			}
		}

		// the explicit type does not implemented any of the allowed merged
		// TypeClasses
		if impl == nil {
			return errors.New(PrintError(2, g, f))
		}
	}

	f.TypeVarMap[v] = w
	g.Atlas[path][funcarg] = w
	return nil
}

// takes information from function f and uses it on g
// to be called when f receives a context C_g
func (f *Function) Update(g *Function) {
	// lock both f and g

	var pf = ConvertPath(f)
	var pgf = ConvertPath(g, f)
	var pg = ConvertPath(g)

	// f is child of g
	if IsChild(f, g) >= 0 {
		// match f() with g(f())
		for funcarg, typevar := range f.Atlas[pf] {
			err := g.updateTypevar(pgf, funcarg, f, typevar)
			if err != nil {
				fmt.Printf(err.Error())
			}
		}

		f.Parents[g] = true
	}

	// replace any type variables that f has replaced elsewhere before
	// this way, type variables "trickle up" the call tree
	for funcarg, typevar := range g.Atlas[pg] {
		if f.TypeVarMap[typevar] != nil {
			err := g.updateTypevar(pg, funcarg, f, f.Atlas[pf][funcarg])
			if err != nil {
				fmt.Printf(err.Error())
			}
		}
	}

	// merge error types: E_g = E_g union E_f
	for errorType := range f.Errors {
		g.Errors[errorType] = true
	}
}

// collects all explicit implementations of f by walking up its call tree
func (f *Function) CollectImplementations(g *Function) (implementations []map[int]*Type, err error) {
	pf := ConvertPath(f)

	// search for explicit types of f's typevars in g
	implementation := make(map[int]*Type)
	for funcarg, typevar := range f.Atlas[pf] {
		if g.TypeMap[typevar] == nil {
			break
		}
		implementation[funcarg] = g.TypeMap[typevar]
	}

	if len(implementation) == len(f.Atlas[pf]) {
		implementations = append(implementations, implementation)
	} else if len(g.Parents) == 0 {
		// we are at a parent but it has no implementation for us?
		// type must be unresolved
		err = errors.New(PrintError(3, f, nil))
	}

	for fun := range g.Parents {
		impl, err := f.CollectImplementations(fun)
		if err != nil {
			return implementations, err
		}
		// append the implementations collected from parents
		for _, t := range impl {
			implementations = append(implementations, t)
		}
	}
	return implementations, err
}

// quickly hacked together print function for one implementation of f
func (f *Function) PrintImplementation(typemap map[int]*Type) {
	fmt.Printf("func %v(", f.Name)
	for i, typ := range typemap {
		if i >= 1 {
			fmt.Printf("%v, ", typ.Name)
		}
	}
	fmt.Printf(") %v \n", typemap[0].Name)
	fmt.Printf("= ")
	if len(f.Children[0]) == 0 {
		fmt.Printf("%v\n", typemap[0].Name)
	} else {
		for g := range f.Children[0] {
			fmt.Printf("%v(...)\n", g.Name)
		}
	}
}

// collect all explicit implementations of f and print to somewhere
func (f *Function) Finish() {
	impl, err := f.CollectImplementations(f)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	for _, typemap := range impl {
		f.PrintImplementation(typemap)
	}
}
