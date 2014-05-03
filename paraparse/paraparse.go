package paraparse

import (
	"github.com/skelterjohn/gopp"
	"Paratype/context"
	"strings"
	"reflect"
	"fmt"
)

var _ = context.TypeClass{}

type Base struct {
	TypeclassDecls []TypeclassImpl
	TypeDecls []Type
	FuncDecls []Func
}

type Typeclass struct {
	Name string
}

type TypeclassImpl struct {
	Name Typeclass
	Inherits []Typeclass
	LastInherit Typeclass
}

type Type struct {
	Name string
	Implements []Typeclass
	LastImplement Typeclass
}

type Func struct {
	Name string
	Arguments []interface{}
	LastArgument interface{}
	ReturnType interface{}
	Constraints []Constraint
	LastConstraint Constraint
	Errors []Error
	LastError Error
	Expr
}

type Expr interface{}

type Error struct {
	Name string
}

type Constraint struct {
	Name TypeVar
	Tclasses []Typeclass
	LastTClass Typeclass
}

type TypeVar struct {
	Name string
}

type TypeName struct {
	Name string
}

type FuncCall struct {
	Name string
	Arguments []interface{}
	LastArgument interface{}
}


const paragopp = `
## ignore comments that begin at beginning of line
ignore: /^#.*\n/
## ignore whitespace at beginning of line
ignore: /^(?:[ \t])+/

Start => {type=Base} {field=TypeclassDecls} <<TypeclassDecl>>* {field=TypeDecls} <<TypeDecl>>* {field=FuncDecls} <<FuncDecl>>+

CommaSep => ','
FuncName => <ident>
TypeclassName => {type=Typeclass} {field=Name} <uident>
TypeVar => {type=TypeVar} {field=Name} <typevar>
TypeName => {type=TypeName} {field=Name} <ident>
ErrorType => {type=Error} {field=Name} <ident>
TypePlace => <TypeVar>
TypePlace => <TypeName>
FuncArgss => <TypePlace> <CommaSep>
FuncArgs => {field=Arguments} <<FuncArgss>>* [{field=LastArgument} <<TypePlace>>]
FuncErrorss => <ErrorType> <CommaSep>
FuncErrors => {field=Errors} <<FuncErrorss>>* [{field=LastError} <<ErrorType>>]
CallArgss => <Expr> <CommaSep>
CallArgs => {field=Arguments} <<CallArgss>>* [{field=LastArgument} <<Expr>>]
Expr => {type=FuncCall} {field=Name} <FuncName> '(' <CallArgs> ')'
Expr => <TypePlace>
TypeDecl => 'type ' <TypeName> ['implements ' {field=Implements} <<TypeClasss>>* {field=LastImplement} <<TypeclassName>>] '\n'
TypeClasss => <TypeclassName> <CommaSep>
TypeclassDecl => 'typeclass ' {field=Name} <<TypeclassName>> ['inherits ' {field=Inherits} <<TypeClasss>>* {field=LastInherit} <<TypeclassName>>] '\n'
FuncConstraint => {type=Constraint} {field=Name} <<TypeVar>> '<' {field=Tclasses} <<TypeClasss>>* {field=LastTClass} <<TypeclassName>> '>'
FuncConstraintss => <FuncConstraint> <CommaSep>
FuncConstraints => 'constrain ' {field=Constraints} <<FuncConstraintss>>* {field=LastConstraint} <<FuncConstraint>>
FuncSig => 'func ' {field=Name} <FuncName> [<FuncConstraints>] '(' <FuncArgs> ')' {field=ReturnType} <<TypePlace>> ['throws ' <FuncErrors>]
FuncDecl => {type=Func} <FuncSig> '\n=' {field=Expr} <<Expr>> '\n'

ident = /([a-z][a-zA-Z]*)/
uident = /([N-Z][a-zA-Z]*)/
typevar = /([A-M][a-zA-Z]*)/
`
func ParseCode() error {
	df, err := gopp.NewDecoderFactory(paragopp, "Start")
	if err != nil {
		fmt.Println(err)
		return err
	}
	df.RegisterType(TypeVar{})
	df.RegisterType(TypeName{})
	df.RegisterType(Func{})
	df.RegisterType(FuncCall{})
	df.RegisterType(Error{})
	df.RegisterType(Typeclass{})
	df.RegisterType(Constraint{})
	//dec := df.NewDecoder(strings.NewReader("typeclass Num\ntype y implements Zun, Num\ntype z implements Num\nfunc foo constrain A <Num, Zun> (d, A, y) iNT throws bigError, gError\n=x\n"))
	dec := df.NewDecoder(strings.NewReader("typeclass Num inherits Zin\ntypeclass Zin\ntype z implements Zin\nfunc foo(d, A) iNT\n=x\n"))
	out := &Base{}
	err = dec.Decode(out)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("%#v\n", out)
	fmt.Println(reflect.TypeOf(out.FuncDecls[0].LastError))

	fmt.Printf("\n")
	tclist, tlist, err := ParseTypeClassDecls(out)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("%+v\n", tlist)
	fmt.Printf("%+v\n", tclist)

	return err
}

// Once a paratype source file has been parsed it is contained in an object
// of type base. This type contains a list that contains all of the type
// classes in the source file and information about what they inherit. This
// function will parse that list and place the resulting output into a
// TypeClass object from the context package. This enables the paratype
// type checker.
func ParseTypeClassDecls(data *Base) ([]context.TypeClass, []context.Type, error) {
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

	TypeSlice := make([]context.Type, len(data.TypeDecls))

	for i, elem := range data.TypeDecls {
		TypeSlice[i].Name =elem.Name
		TypeSlice[i].Implements = make(map[*context.TypeClass]bool)
		for _, implemented := range elem.Implements {
			i_ref, ok := ReferenceMap[implemented.Name]
			if !ok {
				return nil, nil, fmt.Errorf(
					"ParseTypeDecl: TypeClass %s does not exist.",
					implemented.Name)
			}
			TypeSlice[i].Implements[i_ref] = true
		}
		if elem.LastImplement.Name != "" {
			i_ref, ok := ReferenceMap[elem.LastImplement.Name]
			if !ok {
				return nil, nil, fmt.Errorf(
					"ParseTypeDecl: Typeclass %s does not exist.",
					elem.LastImplement.Name)
			}
			TypeSlice[i].Implements[i_ref] = true
		}
		TypeSlice[i].Implements[nil] = true
	}
	return TypeClassSlice, TypeSlice, nil
}

