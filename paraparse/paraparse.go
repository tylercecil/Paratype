package paraparse

import (
	"Paratype/context"
	"fmt"
)

func Setup(code string) ([]context.TypeClass, []context.Type, []context.Function, error) {
	out, err := ParseCode(code)
	if err != nil {
		return nil, nil, nil, err
	}
	fmt.Printf("BASE: %+v\n\n", out)
	TypeClassSlice, ReferenceMap, err := ParseTypeClassDecls(out)
	if err != nil {
		return nil, nil, nil, err
	}
	TypeSlice, err := ParseTypeDecls(out, ReferenceMap)
	if err != nil {
		return nil, nil, nil, err
	}
	FuncSlice, err := ParseFuncDecls(out)
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

// This is responsible for taking the output of the ParseCode function and
// pulling out the type declarations and filling out a type slice.
func ParseTypeDecls(data *Base, ReferenceMap map[string]*context.TypeClass) ([]context.Type, error) {
	TypeSlice := make([]context.Type, len(data.TypeDecls))

	for i, elem := range data.TypeDecls {
		TypeSlice[i].Name = elem.Name
		TypeSlice[i].Implements = make(map[*context.TypeClass]bool)
		for _, implemented := range elem.Implements {
			i_ref, ok := ReferenceMap[implemented.Name]
			if !ok {
				return nil, fmt.Errorf(
					"ParseTypeDecl: TypeClass %s does not exist.",
					implemented.Name)
			}
			TypeSlice[i].Implements[i_ref] = true
		}
		if elem.LastImplement.Name != "" {
			i_ref, ok := ReferenceMap[elem.LastImplement.Name]
			if !ok {
				return nil, fmt.Errorf(
					"ParseTypeDecl: Typeclass %s does not exist.",
					elem.LastImplement.Name)
			}
			TypeSlice[i].Implements[i_ref] = true
		}
		TypeSlice[i].Implements[nil] = true
	}

	return TypeSlice, nil
}

func ParseFuncDecls(data *Base) ([]context.Function, error) {
	FuncSlice := make([]context.Function, len(data.FuncDecls))

	for i, elem := range data.FuncDecls {
		FuncSlice[i].Errors = make(map[*context.Type]bool)
		FuncSlice[i].Name = elem.Name
		FuncSlice[i].Id = i
		FuncSlice[i].NumArgs = len(elem.Arguments) + 2
		for _, errorT := range elem.Errors {
			if errorT.Name != "" {
				v := new(context.Type)
				v.Name = errorT.Name
				FuncSlice[i].Errors[v] = true
				fmt.Printf("\n\nNAME: %s\n\n", v.Name)
			}
		}
		if elem.LastError.Name != "" {
			v := new(context.Type)
			v.Name = elem.LastError.Name
			FuncSlice[i].Errors[v] = true
			fmt.Printf("\n\nNAME: %s\n\n", v.Name)
		}

	}

	return FuncSlice, nil
}

