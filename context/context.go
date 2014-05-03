package context

type TypeClass struct {
	name		string
	inherits	map[*TypeClass]bool
}

var AnyTC TypeClass = TypeClass{name : "ANY"}

// Representation of a Type-Variable in code.
type TypeVariable struct {
	// for ease of access, maybe flatten the hierarchy of typeclasses here?
	constraints map[*TypeClass]bool
	resolved	bool
	name		string
};

// Representation of a specific type in code (as in int, float, ect...)
type Type struct {
	name		string
	implements	map[*TypeClass]bool
};

// a Type struct with name "" is the incomplete type -- all incompletes will
// refer to this
var Incomplete Type = Type{name: ""}

//Representation of a "Function Actor", the main component of Paratype.
type Function struct {
	name        string
	args        []FunctionArg
	Context
}

type Path []struct {
	function	*Function
	cycleNum	int // f^(n) per our notation
}

//A Context object represents information about the implementation of
//a function, and its relationship to other functions.
type Context struct {
	atlas		map[*Path](map[*FunctionArg]*TypeVariable)
	typeMap		map[*TypeVariable]*Type
	typeVarMap	map[*TypeVariable]*TypeVariable
	errors		map[*Type]bool
	children	map[*Context]bool
	parents		map[*Context]bool
}

//FunctionArg structs are used to represent function arguments in an atlas.
//For example, `func f(int x, int y) int` has three FunctionArg's. Position
//may not be necessary, as FunctionArgs are already stored as an array.
type FunctionArg struct {
	function *Function
	position int
}
