package context

import (
	"bytes"
	"strconv"
	"fmt"
	"errors"
	"strings"
	"os"
)

// error messages that merging may throw
var errorMsgs = [...]string{
	"Explicit type conflict",
	"Type class conflict",
	"Explicit type not of merged type class",
	"Missing implementation",
};

// to assist with debugging: print type map, atlas, and type variable map of f
func PrintAll(f *Function) {
	fmt.Printf("\nTypemap of %v\n", f.Name)
	PrintTypeMap(f.TypeMap)
	fmt.Printf("\nAtlas of %v\n", f.Name)
	PrintAtlas(f)
	fmt.Printf("\nTypevarmap of %v\n", f.Name)
	PrintTypeVarMap(f)
}

// to assist with debugging
func PrintTypeMap(typemap map[*TypeVariable]*Type) {
	for tv, t := range typemap {
		fmt.Printf("%+v : %+v\n", tv, t)
	}
}

// to assist with debugging
func PrintAtlas(g *Function) {
	for path, tuple := range g.Atlas {
		fmt.Printf("%+v\n", path)
		for _, tv := range tuple {
			fmt.Printf("%+v\n", tv)
		}
	}
}

// to assist with debugging
func PrintTypeVarMap(g *Function) {
	for tv, tvf := range g.TypeVarMap {
		fmt.Printf("%v to %v\n", tv.Name, tvf.Name)
	}
}

// helper function for creating type error strings
// to add later: print what exactly in the function threw the type error
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
// very annoying way of representing paths, but the easiest one at the moment
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

// converts a path that is a string back to an array of function pointers
func PathToFunctions(path string, allfuncs map[*Function]bool) []*Function {
	ids := strings.Split(path, "-")
	funcs := make([]*Function, len(ids))
	for i, stringid := range ids {
		id, err := strconv.Atoi(stringid)
		if err != nil {
			// we assume it's correct for the moment
		}
		for fun := range allfuncs {
			if fun.Id == id {
				funcs[i] = fun
			}
		}
	}
	return funcs
}

// add function f to an existing path
func AddToPath(path string, f *Function) string {
	var buf bytes.Buffer
	buf.WriteString(path)
	buf.WriteString("-")
	buf.WriteString(strconv.Itoa(f.Id))
	return buf.String()
}

// updates typevar v in func g to be typevar w in func f
// really, we are merging v and w to be w in g
// replaces the type variable v in g.Atlas[path][funcarg] with the type
// variable w that is currently in f somewhere
// 1) will try to merge their explicit types (f.TypeMap[w] and g.TypeMap[v])
// 2) will try to merge the type classes associated with v and w (the type
//    classes that constraint v and w)
func (g *Function) updateTypevar(path string, funcarg int, f *Function,
	w *TypeVariable) error {
	v := g.Atlas[path][funcarg]

	// 1) merging explicit types if possible
	if (f.TypeMap[w] != nil && g.TypeMap[w] != nil &&
			f.TypeMap[w] != g.TypeMap[w]) ||
		(f.TypeMap[w] != nil && g.TypeMap[v] != nil && f.TypeMap[w] !=
		g.TypeMap[v]) ||
		(g.TypeMap[v] != nil && g.TypeMap[w] != nil && g.TypeMap[w] !=
		g.TypeMap[v]) {
		// explicit types do not match
		return errors.New(PrintError(0, g, f))
	}

	// make typemap entry for w in g
	// find explicit type if it exists (nil otherwise)
	if g.TypeMap[w] != nil {
		//fmt.Printf("%+v %+v %+v %+v\n", w.Name, g.Name, v.Name, g.TypeMap[w])
		g.TypeMap[v] = g.TypeMap[w]

	} else {
		if g.TypeMap[v] != nil {
			g.TypeMap[w] = g.TypeMap[v]
		} else {
			g.TypeMap[w] = f.TypeMap[w]
		}
	}

	// did we resolve this type?
	if g.TypeMap[w] != nil {
		w.Resolved = true
	}

	// intersection of w.Constraints and v.Constraints, nil acts as superset
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

// UPDATE (MERGE)
// Function actor f will call this to add information to g
// Will:
// 1) Try to match function calls to function declarations
//    If f is a child of g, then g's call to f can be matched to f's
//    declaration and f's type variables will supersede.
//    Both f and g will have a record of which type variables were replaced
//    with which type variables
// 2) Will replace any previously replaced type variables in g's declaration
//    and in other calls that g makes
func (f *Function) Update(g *Function) error {
	// lock both f and g

	g.Lock() // write lock
	defer g.Unlock()
	f.RLock() // read lock
	defer f.RUnlock()

	// DEBUGGING
	// fmt.Printf("read lock %s write lock %s\n", f.Name, g.Name)
	// defer fmt.Printf("releasing read %s write %s\n", f.Name, g.Name)

	var pf = FunctionsToPath(f)
	var pgf = FunctionsToPath(g, f)

	// 1) If f is a child of g, try to match g's call to f with the declaration
	//    of f
	if f.Parents[g] {
		// match f() with g(f())
		for funcarg, typevar := range f.Atlas[pf] {
			err := g.updateTypevar(pgf, funcarg, f, typevar)
			if err != nil {
				return err
			}
		}
	}

	// Replace any type variables that f has replaced elsewhere before.
	// This way, type variables "trickle up" the call tree
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

func FindFinalTypeVar(tv *TypeVariable, g *Function) (*TypeVariable, int) {
	if g.TypeMap[tv] != nil {
		return tv, 0
	} else if g.TypeVarMap[tv] != nil {
		return FindFinalTypeVar(g.TypeVarMap[tv], g)
	} else {
		return tv, -1
	}
}

// collects all explicit implementations of f by walking up its call tree
func (f *Function) CollectImplementations(g *Function) (implementations []map[*TypeVariable]*Type, err error) {
	pf := FunctionsToPath(f)

	// search for explicit types of f's typevars in g
	implementation := make(map[*TypeVariable]*Type)
	noimpl := false
	for _, typevar := range f.Atlas[pf] {
		tt, ok := FindFinalTypeVar(typevar, g)
		if ok == -1 {
			// look for reverse maps
			noimpl = true
			break
		} else {
			implementation[typevar] = g.TypeMap[tt]
		}
	}


	if noimpl == false {
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// print an implementation of f
func (f *Function) PrintImplementation(typemap map[*TypeVariable]*Type, outFile *os.File) {
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

	_, err := fmt.Fprintf(outFile, "func %v(%s) %v", f.Name, strings.Join(s, ", "), typemap[f.Atlas[pf][0]].Name)
	check(err)

	if len(f.Errors) > 0 {
		_, err = fmt.Fprintf(outFile, " throws %v", strings.Join(r, ", "))
		check(err)
	}
	_, err = fmt.Fprintf(outFile, "\n= ")
	check(err)

	if len(f.Children[0]) == 0 {
		_, err = fmt.Fprintf(outFile, "%v\n", typemap[f.Atlas[pf][0]].Name)
		check(err)
	} else {
		for tv, typ := range f.TypeMap {
			if typ != nil {
				typemap[tv] = typ
			}
		}

		for g := range f.Children[0] {
			_, err = fmt.Fprintf(outFile, "%v\n", PrintRecursiveArgument(f, g, 0, typemap))
			check(err)
		}
	}
}

// collect all explicit implementations of f
func (f *Function) Finish() ([]map[*TypeVariable]*Type, error) {
	impl, err := f.CollectImplementations(f)
	return impl, err
}
