package context

type TypeClass struct {
	name		string
	inherits	[]*TypeClass
}

//Representation of a Type-Variable in code.
type TypeVariable struct {
	// for ease of access, maybe flatten the hierarchy of typeclasses here?
	constraints []*TypeClass
	name		string
	//creator		*Function
};

//Representation of a specific type in code (as in int, float, ect...)
type Type struct {
	name		string
	implements	[]*TypeClass
};

//Representation of a "Function Actor", the main component of Paratype.
type Function struct {
	name        string
	rootContext Context
	args        []FunctionArg
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
	errors		[]*Type
	children	[]*Context
	parents		[]*Context
}

//FunctionArg structs are used to represent function arguments in an atlas.
//For example, `func f(int x, int y) int` has three FunctionArg's. Position
//may not be necessary, as FunctionArgs are already stored as an array.
type FunctionArg struct {
	function *Function
	position int
}
