package context

import (
	"sync"
)

// Object to represent a communication
type Communication struct {
	// path that this communication came through
	Path		string

	// function object that the receiving function actor should add info to
	Context		*Function

	// how deep inside function composition is the call that  spawned this 
	// communication object?
	Depth		int

	// is this the last communication from the function actor that sent this 
	// communication?
	LastComm	bool

	// in case of function composition, any function 
	Wait		*sync.WaitGroup
}

// Object to represent type classes
type TypeClass struct {
	// name of type class
	Name		string

	// which type classes does this inherit from?
	Inherits	map[*TypeClass]bool
}

// Representation of a type variable in code.
type TypeVariable struct {
	// name (actually irrelevant, but used for testing and debugging)
	Name		string

	// which type classes is this TV restrained to? (flattened hierarchy)
	Constraints map[*TypeClass]bool

	// is this TV resolved?
	Resolved	bool
};

// Representation of a specific type in code (as in int, float, ect...)
type Type struct {
	// name (relevant since it must be printed back to file)
	Name		string

	// which type classes may this type implement?
	Implements	map[*TypeClass]bool
};

// Representation of a "function actor", the main component of Paratype.
type Function struct {
	// name of the function
	Name				string

	// unique identifier (used for paths)
	Id					int

	// channel that this function may *receive* on
	Channel				chan *Communication

	// to assist in halting: when all my parents are done and I have finished
	// computing and sending messages, I know that I'm done
	NumParentsDone		int

	// to assist in resolving function composition: what "depth" of children am
	// I (the function actor) currently resolving?
	Depth				int

	// to assist in resolving function composition: at Depth, there are
	// currently a WaitChildren amount of unresolved contexts
	// when this is 0, I can move on to resolve the next-higher level of
	// function composition
	WaitChildren		*sync.WaitGroup

	// to assist in halting in the implementation collection step: all function
	// actors have to have collected all implementations of themselves before 
	// the program can finish
	ImplementationWait	*sync.WaitGroup

	// to assist in halting in the type resolution step: when set to 1, no
	// function may use any channels and must abort (used when a type error
	// arises)
	KillFlag			*sync.WaitGroup

	// to assist in halting: should this function actor abort?
	Dead				bool

	// flag to say whether to collect implementations or abort (in case of type
	// errors in other function actors)
	Implement			bool

	// information about implementation in form of Context object
	Context

	// read-write lock
	sync.RWMutex
}

// A Context object represents information about the implementation of a 
// function and its relationship to other functions.
type Context struct {
	// maps paths in the call graph to a tuple of type variables corresponding
	// to the function arguments of the last function in the path
	// e.g. Atlas[ f(g) ] will contain a tuple of type variables corresponding
	// to g's function arguments
	// used to assist with type resolution
	Atlas			map[string](map[int]*TypeVariable)

	// maps type variables to explicit types when known
	TypeMap			map[*TypeVariable]*Type

	// maps type variables to type variables
	// used to in type resolution: the function actor containing this context
	// will replace any type variable that it has replaced previously by using
	// this map (previously replaced typevars are stored here)
	TypeVarMap		map[*TypeVariable]*TypeVariable

	// acts as a list of errors that the function corresponding to this context
	// or any of its children throw
	Errors			map[*Type]bool

	// children of this function indexed by their composition level
	// for example, for func f(T) = g(m(T), h(T)), we would have
	// Children[0][g] = true,
	// Children[1][m] = true,
	// Children[1][h] = true
	// currently does not allow for multiple instances of the same child, but
	// that would be easily fixable given time.
	Children		map[int]map[*Function]bool

	// all functions that call this function
	Parents			map[*Function]bool

	// to assist in the implementation collection step:
	// array of map that maps type variables to explicit types
	Implementations []map[*TypeVariable]*Type

	// any type errors that may arise during the implementation collection step
	TypeError		error
}

