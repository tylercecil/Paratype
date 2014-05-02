package paratest

import (
    "github.com/skelterjohn/gopp"
    "io/ioutil"
    "strings"
    "reflect"
    "testing"
    "fmt"
)

type Body struct {
    TypeclassDecls []TypeclassDecl
    TypeDecls []TypeDecl
    FuncDecls []FuncDecl
}

type TypeclassDecl struct {
    Name string
    Inherits []string
}

type TypeDecl struct {
    Name string
    Implements []string
}

type FuncDecl struct {
    Signature FuncSig
    FuncBody []interface{}
}

type FuncSig struct {
    Name string
    Constraints []Constraint
    Arguments []interface{}
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
}

const paragopp = `
## ignore comments that begin at beginning of line
ignore: /^#.*\n/
## ignore whitespace at beginning of line
ignore: /^(?:[ \t])+/

Start => <TypeclassDecl>* '\n'* <TypeDecl>+ '\n'+ <FuncDecl>+ '\n'

CommaSep => <ws>* ',' <ws>*
TypeName => <lowerletter> <letter>*
ErrorName => <lowerletter> <letter>*
TypeclassName => <upperletter> <letter>*
TypeVar => <upperletter>+
FuncName => <lowerletter> <letter>*
TypePlace => <TypeVar>
TypePlace => <TypeName>
FuncArg => <ws>* <TypePlace> <ws>*
FuncArgs => (<FuncArg> <CommaSep>)* <FuncArg>
FuncConstraint => <ws>* <TypeVar> <ws>+ 'to' <ws>+ (<TypeclassName> <CommaSep>)* <TypeclassName>
FuncConstraints => <ws>+ 'constrain' <ws>+ (<FuncConstraint> <CommaSep>)* <FuncConstraint>
FuncErrors => <ws>+ 'throws' <ws>+ (<ErrorName> <CommaSep>)* <ErrorName>
TypeclassInherit => <ws>+ 'inherits' <ws>+ (<TypeclassName> <CommaSep>)* <TypeclassName>
TypeImplement => <ws>+ 'implements' <ws>+ (<TypeclassName> <CommaSep>)* <TypeclassName>
TypeDecl => 'type' <ws>+ <TypeName> [<TypeImplement>] '\n'+
TypeclassDecl => 'typeclass' <ws>+ <TypeclassName> [<TypeclassInherit>] '\n'+
Expr => <ws>* <TypeName> <ws>*
Expr => <ws>* <TypeVar> <ws>*
Expr => <ws>* <FuncName> <ws>* '(' [(<Expr> <CommaSep>)* <Expr>] ')' <ws>*
FuncSig => 'func' <ws>+ <FuncName> [<FuncConstraints>] <ws>+ '(' <FuncArgs> ')' <ws>+ <TypePlace> [<FuncErrors>]
FuncDecl => <FuncSig> '\n' <ws>* '=' <Expr> '\n'+

lowerletter = /([a-z])/
upperletter = /([A-Z])/
letter = /([A-Za-z])/
ws = /([ \t])/
`
func TestGrammar(t *testing.T) {
    df, err := gopp.NewDecoderFactory(paragopp, "Start")
    if err != nil {
        t.Error(err)
        return
    }
    dec := df.NewDecoder(strings.NewReader("func foo(int x, int y) int\n=bar(y)"))

    fmt.Printf("%v", ast)
}


