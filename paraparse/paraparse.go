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
	dec := df.NewDecoder(strings.NewReader("typeclass Num\ntypeclass Zin\ntype z\nfunc foo(d, A) iNT\n=x\n"))
    out := &Base{}
	err = dec.Decode(out)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("%#v\n", out)
    fmt.Println(reflect.TypeOf(out.FuncDecls[0].LastError))

    fmt.Printf("\n\n\n")
    PrintTypeclassDecls(out)
    fmt.Printf("\n\n\n")
    PrintTypeDecls(out)
    fmt.Printf("\n\n\n")
    PrintFuncDecls(out)



    return err
}

func PrintTypeclassDecls(data *Base) {
    TypeclassList := make([]context.TypeClass, len(data.TypeclassDecls))
    for i, elem := range data.TypeclassDecls {
        TypeclassList[i].name = elem.Name.Name
    }
    fmt.Printf("%+v\n", TypeclassList)
}

func PrintTypeDecls(data *Base) {
    for _, elem := range data.TypeDecls {
        fmt.Printf("%+v\n", elem)
    }
}

func PrintFuncDecls(data *Base) {
    for _, elem := range data.FuncDecls {
        fmt.Printf("%+v\n", elem)
    }
}
