package paratest

import (
    "github.com/skelterjohn/gopp"
    "strings"
    "testing"
    "reflect"
)

type Base struct {
    TypeclassDecls []Typeclass
    TypeDecls []Type
    FuncDecls []Func
}

type Typeclass struct {
    Name string
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

Start => {type=Base} {field=TypeDecls} <<TypeDecl>>* {field=FuncDecls} <<FuncDecl>>+

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
TypeDecl => 'type ' <TypeName> 'implements ' {field=Implements} <<TypeClasss>>* {field=LastImplement} <<TypeclassName>> '\n'
TypeClasss => <TypeclassName> <CommaSep>
FuncConstraint => {type=Constraint} {field=Name} <<TypeVar>> '<' {field=Tclasses} <<TypeClasss>>* {field=LastTClass} <<TypeclassName>> '>'
FuncConstraintss => <FuncConstraint> <CommaSep>
FuncConstraints => 'constrain ' {field=Constraints} <<FuncConstraintss>>* {field=LastConstraint} <<FuncConstraint>>
FuncSig => 'func ' {field=Name} <FuncName> [<FuncConstraints>] '(' <FuncArgs> ')' {field=ReturnType} <<TypePlace>> ['throws ' <FuncErrors>]
FuncDecl => {type=Func} <FuncSig> '\n=' {field=Expr} <<Expr>> '\n'

ident = /([a-z][a-zA-Z]*)/
uident = /([N-Z][a-zA-Z]*)/
typevar = /([A-M][a-zA-Z]*)/
`
func TestGrammar(t *testing.T) {
    df, err := gopp.NewDecoderFactory(paragopp, "Start")
    if err != nil {
        t.Error(err)
        return
    }
	df.RegisterType(TypeVar{})
	df.RegisterType(TypeName{})
	df.RegisterType(Func{})
    df.RegisterType(FuncCall{})
    df.RegisterType(Error{})
    df.RegisterType(Typeclass{})
    df.RegisterType(Constraint{})
    dec := df.NewDecoder(strings.NewReader("type y implements Zun, Num\ntype z implements Num\nfunc foo constrain A <Num, Zun> (d, A, y) iNT throws bigError, gError\n=x\n"))
	out := &Base{}
	err = dec.Decode(out)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", out)
    t.Logf("%+v\n", reflect.TypeOf(out.FuncDecls[0].LastError))
}


