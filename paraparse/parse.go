package paraparse

import (
    "fmt"
    "os"
    "strings"
    "io"
    "github.com/skelterjohn/gopp"
)

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
func Parse(code string, file bool) (*Base, error) {
    df, err := gopp.NewDecoderFactory(paragopp, "Start")
    if err != nil {
        fmt.Println(err)
        return nil, err
    }
    df.RegisterType(TypeVar{})
    df.RegisterType(TypeName{})
    df.RegisterType(Func{})
    df.RegisterType(FuncCall{})
    df.RegisterType(Error{})
    df.RegisterType(Typeclass{})
    df.RegisterType(Constraint{})

    var fcontents io.Reader

    if file == true {
        fcontents, err = os.Open(code)
        if err!= nil {
            return nil, err
        }
        defer fcontents.(*os.File).Close()
    } else {
        fcontents = df.NewDecoder(strings.NewReader(code))
    }

    dec := df.NewDecoder(fcontents)
    out := &Base{}
    err = dec.Decode(out)
    if err != nil {
        fmt.Println(err)
        return nil, err
    }

    return out, nil
}