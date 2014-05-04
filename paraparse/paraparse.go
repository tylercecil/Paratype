package paraparse

import (
	"Paratype/context"
	"fmt"
	"strconv"
)

var tvname int = 0

func Setup(code string) ([]context.TypeClass, []context.Type, []*context.Function, error) {
	out, err := ParseCode(code)
	if err != nil {
		return nil, nil, nil, err
	}
	fmt.Printf("BASE: %+v\n\n", out)
	TypeClassSlice, ReferenceMap, err := ParseTypeClassDecls(out)
	if err != nil {
		return nil, nil, nil, err
	}
	TypeSlice, TypeMap, err := ParseTypeDecls(out, ReferenceMap)
	if err != nil {
		return nil, nil, nil, err
	}
	FuncSlice, err := ParseFuncDecls(out, ReferenceMap, TypeMap)
	if err != nil {
		return nil, nil, nil, err
	}

	return TypeClassSlice, TypeSlice, FuncSlice, nil
}

// Once a paratype source file has been parsed it is contained in an object
// of type base. This type contains a list that contains all of the type
// classes in the source file and information about what they inherit. This
// function will parse that list and place the resulting output into a
// TypeClass object from the context package. This enables the paratype
// type checker.
func ParseTypeClassDecls(data *Base) ([]context.TypeClass, map[string]*context.TypeClass, error) {
	// The type context.TypeClass consists of two items:
	// 		1) A Name
	// 		2) A map of inherited TypeClasses where the key is a pointer to
	// 			the TypeClass object and the value is a boolean representing
	// 			the inheritance status.
	//
	TypeClassSlice := make([]context.TypeClass, len(data.TypeclassDecls))
	ReferenceMap := make(map[string]*context.TypeClass)

	// Copy the Name and fill the ReferenceMap entry with the appropriate
	// pointer. This will be used to set the references later.
	for i, elem := range data.TypeclassDecls {
		TypeClassSlice[i].Name = elem.Name.Name
		ReferenceMap[elem.Name.Name] = &TypeClassSlice[i]
	}

	// Go through the inherited TypeClasses and retrieve their references.
	// The references are then used as a key in the Inherits map contained
	// in the TypeClass struct. If retrieving the entry from the ReferenceMap
	// fails this signifies a malformed source file. (References a TypeClass
	// that does not exist.)
	for i, elem := range data.TypeclassDecls {
		TypeClassSlice[i].Inherits = make(map[*context.TypeClass]bool)
		for _, inherited := range elem.Inherits {
			i_ref, ok := ReferenceMap[inherited.Name]
			if !ok {
				return nil, nil, fmt.Errorf(
					"ParseTypeClassDecls: TypeClass %s does not exist.",
					 inherited.Name)
			}
			TypeClassSlice[i].Inherits[i_ref] = true
		}
		if elem.LastInherit.Name != "" {
			i_ref, ok := ReferenceMap[elem.LastInherit.Name]
			if !ok {
				return nil, nil, fmt.Errorf(
					"ParseTypeClassDecls: TypeClass %s does not exist.",
					elem.LastInherit.Name)
			}
			TypeClassSlice[i].Inherits[i_ref] = true
		}
		// This is for convention purposes. This will become more relevant
		// when parsing types. Nil represents the basetype.
		TypeClassSlice[i].Inherits[nil] = true
	}

	return TypeClassSlice, ReferenceMap, nil
}

// Given the implementation and the implementation map will assign
// the given implementation to the map with the correct reference to the
// typeclass.8
func AssignImplementation(impl Typeclass, implMap map[*context.TypeClass]bool, ReferenceMap map[string]*context.TypeClass) error {
	if impl.Name != "" {
		i_ref, ok := ReferenceMap[impl.Name]
		if !ok {
			return fmt.Errorf("ParseTypeDecl: TypeClass %s does not exist.",
				impl.Name)
		}
		implMap[i_ref] = true
	}
	return nil
}

// This is responsible for taking the output of the ParseCode function and
// pulling out the type declarations and filling out a type slice.
func ParseTypeDecls(data *Base, ReferenceMap map[string]*context.TypeClass) ([]context.Type, map[string]*context.Type, error) {
	TypeSlice := make([]context.Type, len(data.TypeDecls))
	TypeMap := make(map[string]*context.Type)

	for i, elem := range data.TypeDecls {
		TypeSlice[i].Name = elem.Name
		TypeMap[elem.Name] = &TypeSlice[i]
		TypeSlice[i].Implements = make(map[*context.TypeClass]bool)
		for _, implemented := range elem.Implements {
			if err := AssignImplementation(implemented, TypeSlice[i].Implements, ReferenceMap); err != nil {
				return nil, nil, err
			}
		}
		if err := AssignImplementation(elem.LastImplement, TypeSlice[i].Implements, ReferenceMap); err != nil {
			return nil, nil, err
		}
		TypeSlice[i].Implements[nil] = true
	}
	return TypeSlice, TypeMap, nil
}

// Given an error name and a error map will add the error to the map.
func AssignError(err Error, errMap map[*context.Type]bool) {
	if err.Name != "" {
		v := new(context.Type)
		v.Name = err.Name
		errMap[v] = true
	}
}

func ResolveTypeVar(atlas map[string]map[int]*context.TypeVariable,
	path string,
	ftypemap map[*context.TypeVariable]*context.Type,
	typeVarRef map[string]*context.TypeVariable,
	typeMap map[string]*context.Type,
	pos int,
	arg interface{}) {

	if arg != nil {
		switch arg.(type) {
		case TypeName:
			T := MakeTypeVar(true)
			ftypemap[T] = typeMap[arg.(TypeName).Name]
			atlas[path][pos] = T

		case TypeVar:
			name := arg.(TypeVar).Name
			v, ok := typeVarRef[name]
			if ok {
				atlas[path][pos] = v
			} else {
				T := MakeTypeVar(false)
				ftypemap[T] = nil
				atlas[path][pos] = T
				typeVarRef[name] = T
			}
		}
	}
}

func HandleFuncComposition(ReferenceMap map[string]*context.Function,
	f *context.Function,
	returnVar *context.TypeVariable,
	expr FuncCall,
	typeVarRef map[string]*context.TypeVariable,
	typeMap map[string]*context.Type,
	level int,
) {
	g := ReferenceMap[expr.Name]
	if len(f.Children[level]) == 0 {
		f.Children[level] = make(map[*context.Function]bool)
	}
	f.Children[level][g] = true
	g.Parents[f] = true
	pfg := context.FunctionsToPath(f, g)
	f.Atlas[pfg] = make(map[int]*context.TypeVariable)
	f.Atlas[pfg][0] = returnVar
	expr.Arguments = append(expr.Arguments, expr.LastArgument)

	for pos, arg := range expr.Arguments {
		switch arg.(type) {
		case FuncCall:
			// make return typevar, pass to self
			T := MakeTypeVar(false)
			f.TypeMap[T] = nil
			HandleFuncComposition(ReferenceMap, f, T, arg.(FuncCall), typeVarRef, typeMap, level+1)

		case TypeName:
			// make new typevar
			ResolveTypeVar(f.Atlas, pfg, f.TypeMap, typeVarRef, typeMap, pos+1, arg)

		case TypeVar:
			// match typeVarRef
			ResolveTypeVar(f.Atlas, pfg, f.TypeMap, typeVarRef, typeMap, pos+1, arg)
		}
	}

}

// Will go through the list and parse the function declarations out and into
// A slice of function objects.
func ParseFuncDecls(data *Base, typeClassMap map[string]*context.TypeClass, typeMap map[string]*context.Type) ([]*context.Function, error) {
	FuncSlice := make([]*context.Function, len(data.FuncDecls))
	ReferenceMap := make(map[string]*context.Function)

	for i, elem := range data.FuncDecls {
		FuncSlice[i] = new(context.Function)
		FuncSlice[i].Errors = make(map[*context.Type]bool)
		FuncSlice[i].Name = elem.Name

		// Reference map is needed later for the FindChildren function.
		ReferenceMap[elem.Name] = FuncSlice[i]
		FuncSlice[i].Parents = make(map[*context.Function]bool)
		FuncSlice[i].Id = i
		FuncSlice[i].NumArgs = len(elem.Arguments) + 2

		// AssignError will fill out the Errors field
		for _, errorT := range elem.Errors {
			AssignError(errorT, FuncSlice[i].Errors)
		}
		AssignError(elem.LastError, FuncSlice[i].Errors)
	}
	// Once the ReferenceMap has been built it is important to go through
	// the function declarations and enumerate the children and their depth.

	for _, elem := range data.FuncDecls {
		f := ReferenceMap[elem.Name]
		pf := context.FunctionsToPath(f)
		f.Atlas = make(map[string]map[int]*context.TypeVariable)
		f.Children = make(map[int]map[*context.Function]bool)
		f.TypeVarMap = make(map[*context.TypeVariable]*context.TypeVariable)
		f.TypeMap = make(map[*context.TypeVariable]*context.Type)
		f.Atlas[pf] = make(map[int]*context.TypeVariable)
		typeVarRef := make(map[string]*context.TypeVariable)

		ResolveTypeVar(f.Atlas, pf, f.TypeMap, typeVarRef, typeMap, 0, elem.ReturnType)

		elem.Arguments = append(elem.Arguments, elem.LastArgument)
		for pos, arg := range elem.Arguments {
			ResolveTypeVar(f.Atlas, pf, f.TypeMap, typeVarRef, typeMap, pos+1, arg)
		}

		switch elem.Expr.(type) {
		case FuncCall:
			HandleFuncComposition(ReferenceMap, f, f.Atlas[pf][0], elem.Expr.(FuncCall), typeVarRef, typeMap, 0)

		case TypeName:
		case TypeVar:
		}
	}

	return FuncSlice, nil
}

func MakeTypeVar(res bool) *context.TypeVariable {
	s := new(context.TypeVariable)
	s.Constraints = make(map[*context.TypeClass]bool)
	s.Constraints[nil] = true
	s.Resolved = res
	s.Name = strconv.Itoa(tvname)
	tvname++
	return s
}

func GetName(elem interface{}) string {
	switch elem.(type) {
	case TypeVar:
		return elem.(TypeVar).Name
	case TypeName:
		return elem.(TypeName).Name
	}
	return ""
}

func GetNewTypeVar(Ref map[string]*context.TypeVariable, elem interface{}, TypeCount *int) *context.TypeVariable {
	v, ok := Ref[GetName(elem)]
	if !ok {
		v = new(context.TypeVariable)
		v.Name = fmt.Sprintf("A%d", *TypeCount)
		*(TypeCount)++
		Ref[GetName(elem)] = v
	}
	return v
}

func PerformFill(Ref map[string]*context.TypeVariable, elem interface{}, TypeCount *int, ArgCount *int, Name string, fun *context.Function) {
	v := GetNewTypeVar(Ref, elem, TypeCount)
	fun.Atlas[Name][*ArgCount] = v
	(*ArgCount)++
}

func FillAtlasTypes(fun *context.Function, elem Func) {
	TypeVarRef := make(map[string]*context.TypeVariable)
	fun.Atlas = make(map[string]map[int]*context.TypeVariable)
	Name := context.FunctionsToPath(fun)
	fun.Atlas[Name] = make(map[int]*context.TypeVariable)

	TypeCount := 0
	ArgCount := 0

	PerformFill(TypeVarRef, elem.ReturnType, &TypeCount, &ArgCount, Name, fun)
	for _, arg := range elem.Arguments {
		PerformFill(TypeVarRef, arg, &TypeCount, &ArgCount, Name, fun)
	}
	if n := GetName(elem.LastArgument); n != "" {
		PerformFill(TypeVarRef, elem.LastArgument, &TypeCount, &ArgCount, Name, fun)
	}

	// SOMETHING WAS EVENTUALLY GOING TO HAPPEN HERE
	// for depth, childmap := range fun.Children {
	// 	for child, _ := range childmap {
	// 		Name := context.FunctionsToPath(fun, child)
	// 		if depth == 0 {
	// 			PerformFill(TypeVarRef, elem.ReturnType, &TypeCount, &ArgCount, Name, fun)
	// 		}
	// 	}
	// }
}

// Given the expression for each function call will traverse down the
// function call tree. If expression is a FuncCall then it assigns it to the
// children map. If it is a TypeVar or TypeName then it doesn't matter.
// The parameters of the function are then checked to see if there is any
// composition. Also adds the childs parent to its parents list.
func FindChildren(par *context.Function, e Expr, depth int, cMap map[int]map[*context.Function]bool, rMap map[string]*context.Function) error {
	// Type switch over the expression given.
	switch e.(type) {
	// If the expression given is a function call then the function is added
	// to the children map. It makes sure the function exists and errors if
	// it does not.
	case FuncCall:
		// Get the pointer to the function object
		FuncP, ok := rMap[e.(FuncCall).Name]
		if !ok {
			return fmt.Errorf("FindChildren: Function %s does not exist.",
				e.(FuncCall).Name)
		}
		FuncP.Parents[par] = true
		cMap[depth] = make(map[*context.Function]bool)
		cMap[depth][FuncP] = true
		// Go through all of the arguments checking those expressions
		for _, elem := range e.(FuncCall).Arguments {
			err := FindChildren(par, elem, depth+1, cMap, rMap)
			if err != nil {
				return err
			}
		}
		// Make sure to also check the last argument. This may not exist.
		if e.(FuncCall).LastArgument != nil {
			err := FindChildren(par, e.(FuncCall).LastArgument, depth+1, cMap, rMap)
			if err != nil {
				return err
			}
		}
	case TypeVar:
	case TypeName:
	}

	return nil
}

