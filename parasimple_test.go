package paratest

import (
    "github.com/skelterjohn/gopp"
    "strings"
    "testing"
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

type TypePlace struct{
}

type Func struct {
    Name string
	ReturnType TypePlace
}

type TypeVar struct {
	TypePlace
	Name string
}

type TypeName struct {
	TypePlace
	Name string
}

const paragopp = `
## ignore comments that begin at beginning of line
ignore: /^#.*\n/
## ignore whitespace at beginning of line
ignore: /^(?:[ \t])+/

Start => {type=Base} {field=FuncDecls} <<FuncDecl>>+

FuncName => <ident>
TypeVar => {type=TypeVar} {field=Name} <typevar>
TypeName => {type=TypeName} {field=Name} <ident>
TypePlace => <TypeVar>
TypePlace => <TypeName>
FuncDecl => 'func' <ws>+ {field=Name} <FuncName> {field=ReturnType} <<TypePlace>>

ident = /([a-z][a-zA-Z]*)/
typevar = /([A-Z]*)/
ws = /([ \t])/
`
func TestGrammar(t *testing.T) {
    df, err := gopp.NewDecoderFactory(paragopp, "Start")
    if err != nil {
        t.Error(err)
        return
    }
    dec := df.NewDecoder(strings.NewReader("func foo int"))
	out := &Base{}
	err = dec.Decode(out)
	if err != nil {
		t.Error(err)
	}
	t.Logf("name %v\n", out.FuncDecls)
    t.Logf("abc %T", dec)
}


