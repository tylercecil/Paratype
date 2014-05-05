package context

import (
	"bytes"
	"strconv"
	"fmt"
	"errors"
	"strings"
)

var errorMsgs = [...]string{
	"Explicit type conflict",
	"Type class conflict",
	"Explicit type not of merged type class",
	"Missing implementation",
};

func PrintAll(f *Function) {
	fmt.Printf("\nTypemap of %v\n", f.Name)
	PrintTypeMap(f)
	fmt.Printf("\nAtlas of %v\n", f.Name)
	PrintAtlas(f)
	fmt.Printf("\nTypevarmap of %v\n", f.Name)
	PrintTypeVarMap(f)

}

func PrintTypeMap(g *Function) {
	for tv, t := range g.TypeMap {
		fmt.Printf("%+v : %+v\n", tv, t)
	}
}

func PrintAtlas(g *Function) {
	for path, tuple := range g.Atlas {
		fmt.Printf("%+v\n", path)
		for _, tv := range tuple {
			fmt.Printf("%+v\n", tv)
		}
	}
}

func PrintTypeVarMap(g *Function) {
	for tv, tvf := range g.TypeVarMap {
		fmt.Printf("%v to %v\n", tv.Name, tvf.Name)
	}
}

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
func FunctionsToPath(f ...interface{}) string {
	var buf bytes.Buffer

	for _, fun := range f {
		buf.WriteString(strconv.Itoa(fun.(*Function).Id))
		buf.WriteString("-")
	}
	s := buf.String()
	s = s[:len(s)-1] // remove last character
	return s
}

func PathToFunctions(path string, allfuncs map[*Function]bool) []*Function {
	ids := strings.Split(path, "-")
	funcs := make([]*Function, len(ids))
	for i, stringid := range ids {
		id, err := strconv.Atoi(stringid)
		if err != nil {
			// 
		}
		for fun := range allfuncs {
			if fun.Id == id {
				funcs[i] = fun
			}
		}
	}
	return funcs
}

// add function f to path
func AddToPath(path string, f *Function) string {
	var buf bytes.Buffer
	buf.WriteString(path)
	buf.WriteString("-")
	buf.WriteString(strconv.Itoa(f.Id))
	return buf.String()
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
	g.TypeVarMap[v] = w
	g.Atlas[path][funcarg] = w
	return nil
}

// takes information from function f and uses it on g
// to be called when f receives a context C_g
func (f *Function) Update(g *Function) error {
	// lock both f and g

	// DEADLOCK POSSIBLE FOR CYCLES
	g.Lock() // write lock
	defer g.Unlock()
	f.RLock() // read lock
	defer f.RUnlock()

	fmt.Printf("read lock %s write lock %s\n", f.Name, g.Name)
	defer fmt.Printf("releasing read %s write %s\n", f.Name, g.Name)

	var pf = FunctionsToPath(f)
	var pgf = FunctionsToPath(g, f)

	// f is child of g
	if f.Parents[g] {
		// match f() with g(f())
		for funcarg, typevar := range f.Atlas[pf] {
			err := g.updateTypevar(pgf, funcarg, f, typevar)
			if err != nil {
				return err
			}
		}
	}

	// replace any type variables that f has replaced elsewhere before
	// this way, type variables "trickle up" the call tree
	for path, atlasentry := range g.Atlas {
		for funcarg, typevar := range atlasentry {
			if f.TypeVarMap[typevar] != nil {
				err := g.updateTypevar(path, funcarg, f, f.TypeVarMap[typevar])
				if err != nil {
					return err
				}
			}
		}
	}

	// merge error types: E_g = E_g union E_f
	for errorType := range f.Errors {
		g.Errors[errorType] = true
	}
	return nil
}

func FindFinalTypeVar(tv *TypeVariable, g *Function) *TypeVariable {
	if g.TypeMap[tv] != nil {
		return tv
	} else if g.TypeVarMap[tv] != nil {
		return FindFinalTypeVar(g.TypeVarMap[tv], g)
	} else {
		return nil
	}
}

// collects all explicit implementations of f by walking up its call tree
func (f *Function) CollectImplementations(g *Function) (implementations []map[*TypeVariable]*Type, err error) {
	pf := FunctionsToPath(f)

	// search for explicit types of f's typevars in g
	implementation := make(map[*TypeVariable]*Type)
	for _, typevar := range f.Atlas[pf] {
		tt := FindFinalTypeVar(typevar, g)
		if tt == nil {
			break
		} else {
			implementation[typevar] = g.TypeMap[tt]
		}
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

func PrintRecursiveArgument(
	f *Function,
	g *Function,
	level int,
	tmap map[*TypeVariable]*Type) string {

	var pfg = FunctionsToPath(f, g)

	var r []string = make([]string, len(f.Atlas[pfg])-1)
	for i, typevar := range f.Atlas[pfg] {
		if i >= 1 {
			// find composed calls
			usetype := true
			for h := range f.Children[level+1] {
				var ph = FunctionsToPath(f, h)
				// return type matching? has to be in this place
				if f.Atlas[ph][0] == typevar {
					r[i-1] = PrintRecursiveArgument(f, h, level+1, tmap)
					usetype = false
					break
				}
			}
			if usetype {
				r[i-1] = tmap[typevar].Name
			}
		}
	}
	return fmt.Sprintf("%v(%v)", g.Name, strings.Join(r, ", "))
}

// quickly hacked together print function for one implementation of f
func (f *Function) PrintImplementation(typemap map[*TypeVariable]*Type) {
	s := make([]string, len(typemap)-1)
	i := 0
	for _, typ := range typemap {
		if i >= 1 {
			s[i-1] = typ.Name
		}
		i++
	}

	r := make([]string, len(f.Errors))
	i = 0
	for e := range f.Errors {
		r[i] = e.Name
		i++
	}
	var pf = FunctionsToPath(f)

	fmt.Printf("func %v(%s) %v", f.Name, strings.Join(s, ", "), typemap[f.Atlas[pf][0]].Name)

	if len(f.Errors) > 0 {
		fmt.Printf(" throws %v", strings.Join(r, ", "))
	}
	fmt.Printf("\n= ")

	if len(f.Children[0]) == 0 {
		fmt.Printf("%v\n", typemap[f.Atlas[pf][0]].Name)
	} else {
		for tv, typ := range f.TypeMap {
			typemap[tv] = typ
		}

		for g := range f.Children[0] {
			fmt.Printf("%v\n", PrintRecursiveArgument(f, g, 0, typemap))
		}
	}
}

// collect all explicit implementations of f and print to somewhere
func (f *Function) Finish() ([]map[*TypeVariable]*Type, error) {
	impl, err := f.CollectImplementations(f)
	return impl, err
	/*if err != nil {
		fmt.Printf(err.Error())
		return
	}

	for _, typemap := range impl {
		f.PrintImplementation(typemap)
	}*/
}
