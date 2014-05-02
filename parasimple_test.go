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
}

type Func struct {
    Name string
	Arguments []interface{}
	LastArgument interface{}
	ReturnType interface{}
	Errors []Error
	LastError interface{}
	Expr
}

type Expr interface{}

type Error struct {
	Name interface{}
}

type TypeVar struct {
	Name string
}

type TypeName struct {
	Name string
}

type FuncCall struct {
    Name string
    Parameters []interface{}
    LastParameter interface{}
}


const paragopp = `
## ignore comments that begin at beginning of line
ignore: /^#.*\n/
## ignore whitespace at beginning of line
ignore: /^(?:[ \t])+/

Start => {type=Base} {field=FuncDecls} <<FuncDecl>>+

CommaSep => ','
FuncName => <ident>
TypeVar => {type=TypeVar} {field=Name} <typevar>
TypeName => {type=TypeName} {field=Name} <ident>
TypePlace => <TypeVar>
TypePlace => <TypeName>
FuncArgss => <TypePlace> <CommaSep>
FuncArgs => {field=Arguments} <<FuncArgss>>* [{field=LastArgument} <<TypePlace>>]
FuncErrorss => {field=Name} <TypePlace> <CommaSep>
FuncErrors => {field=Errors} <<FuncArgss>>* {field=LastError} <TypePlace>
Expr => <TypePlace>
Expr => {type=FuncCall} {field=Name} <FuncName> '(' [(<Expr> <CommaSep>)* <Expr>] ')'
FuncSig => 'func ' {field=Name} <FuncName> '(' <FuncArgs> ')' {field=ReturnType} <<TypePlace>> ['throws ' <FuncErrors>]
FuncDecl => {type=Func} <FuncSig> '\n=' {field=Expr} <<Expr>> '\n'

ident = /([a-z][a-zA-Z]*)/
typevar = /([A-Z]*)/
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
    dec := df.NewDecoder(strings.NewReader("func foo(d, g, y) iNT\n=bar\n"))
	out := &Base{}
	err = dec.Decode(out)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", out)
    t.Logf("%+v\n", reflect.TypeOf(out.FuncDecls[0].LastArgument))
}


