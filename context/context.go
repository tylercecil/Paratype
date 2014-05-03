package context

import "sync"

type TypeClass struct {
	Name		string
	Inherits	map[*TypeClass]bool
}

// Representation of a Type-Variable in code.
type TypeVariable struct {
	// for ease of access, maybe flatten the hierarchy of typeclasses here?
	Constraints map[*TypeClass]bool
	Resolved	bool
	Name		string
};

// Representation of a specific type in code (as in int, float, ect...)
type Type struct {
	Name		string
	Implements	map[*TypeClass]bool
};

//Representation of a "Function Actor", the main component of Paratype.
type Function struct {
	Lock		*sync.RWMutex
	Name        string
	Id			int
	NumArgs		int
	Context
}

//A Context object represents information about the implementation of
//a function, and its relationship to other functions.
type Context struct {
	Atlas		map[string](map[int]*TypeVariable) // path -> funcarg -> typevar
	TypeMap		map[*TypeVariable]*Type
	TypeVarMap	map[*TypeVariable]*TypeVariable
	Errors		map[*Type]bool
	Children	map[*Function]bool
	Parents		map[*Function]bool
}

