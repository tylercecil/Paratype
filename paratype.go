//The Main package for Paratype type analysis software.
package main

import "fmt"

//Representation of a Type-Variable in code.
type TypeVariable int;
//Representation of a specific type in code (as in int, float, ect...)
type Type int;

//Representation of a "Function Actor", the main component of Paratype.
type Function struct {
  name        string
  rootContext Context
  args        []FunctionArg
}

//A Context object represents information about the implementation of
//a function, and its relationship to other functions.
type Context struct {
  atlas      map[FunctionArg]TypeVariable
  typeMap    map[TypeVariable]Type
  typeVarMap map[TypeVariable]TypeVariable
  children   []Context
  parents    []Context
}

//FunctionArg structs are used to represent function arguments in an atlas.
//For example, `func f(int x, int y) int` has three FunctionArg's. Position
//may not be necessary, as FunctionArgs are already stored as an array.
type FunctionArg struct {
  function Function
  position int
}

//Dummy main function.
func main() {
  fmt.Println("Paratype")
}
