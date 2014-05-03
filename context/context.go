package context

type TypeClass struct {
	name		string
	inherits	map[*TypeClass]bool
}

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

//Representation of a "Function Actor", the main component of Paratype.
type Function struct {
	name        string
	id			int
	numArgs		int
	Context
}

//type Path string;
/*type Path []struct {
	function	*Function
	cycleNum	int // f^(n) per our notation
}*/

//A Context object represents information about the implementation of
//a function, and its relationship to other functions.
type Context struct {
	atlas		map[string](map[int]*TypeVariable) // path -> funcarg -> typevar
	typeMap		map[*TypeVariable]*Type
	typeVarMap	map[*TypeVariable]*TypeVariable
	errors		map[*Type]bool
	children	map[*Context]bool
	parents		map[*Context]bool
}

