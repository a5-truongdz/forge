package main

import "fmt"

type Parser struct {
    lexer *Lexer
    current Token
    peek Token
}

type ParserError struct {
    Filename string
    Line     int
    Column   int
    Message  string
}

func (e ParserError) Error() string {
    return fmt.Sprintf(
        "%s:%d:%d: %s",
        e.Filename,
        e.Line,
        e.Column,
        e.Message,
    )
}

func NewParser(lexer *Lexer) *Parser {
    p := &Parser {
        lexer: lexer,
    }

    p.advance()
    p.advance()

    return p
}

func (p *Parser) advance() {
    p.current = p.peek
    p.peek = p.lexer.Next()
}

func (p *Parser) checkCurrent(tok TokenEnum) bool {
    return p.current.Type == tok
}

func (p *Parser) checkPeek(tok TokenEnum) bool {
    return p.peek.Type == tok
}

func (p *Parser) error(message string) {
    panic(ParserError{
        Filename: p.lexer.filename,
        Line:     p.current.Line,
        Column:   p.current.Column,
        Message:  message,
    })
}

func (p *Parser) expect(tok TokenEnum) Token {
    if p.current.Type != tok {
        p.error(fmt.Sprintf("expected %s, got %s", tok, p.current.Lexeme))
    }

    token := p.current
    p.advance()
    return token
}

type Node interface {
    node()
}
type Stmt interface {
    Node
    stmt()
}
type Expr interface {
    Node
    expr()
}
type Type interface {
    Node
    typ()
}

type Program struct {
    Statements []Stmt
}
func (Program) node() {}

type BindingStmt struct {
    Name string
    Type Type
    Value Expr
    Mutable bool
}
func (BindingStmt) node() {}
func (BindingStmt) stmt() {}

type IntLiteral struct {
    Value int64
}
func (IntLiteral) node() {}
func (IntLiteral) expr() {}

type IntType struct {}
func (IntType) node() {}
func (IntType) typ() {}

func (p *Parser) parseType() Type {
    switch p.current.Type {
    case TokenInt:
        p.advance()
        return IntType{}

    default:
        p.error(fmt.Sprintf("expected type, got %s", p.current.Lexeme))
        return nil
    }
}

func (p *Parser) parseExpression() Expr {
    switch p.current.Type {
    case TokenLInt:
        token := p.current
        p.advance()

        var value int64
        if _, err := fmt.Sscanf(token.Lexeme, "%d", &value); err != nil {
            p.error("invalid integer literal")
        }

        return IntLiteral{
            Value: value,
        }

    default:
        p.error(fmt.Sprintf(
            "expected expression, got %s",
            p.current.Lexeme,
        ))
        return nil
    }
}

func (p *Parser) parseBinding() Stmt {
    name := p.expect(TokenIdentifier)

    var typ Type
    if p.checkCurrent(TokenColon) {
        p.advance()

        typ = p.parseType()
        p.expect(TokenEqual)
    } else {
        p.expect(TokenDeclare)
    }

    value := p.parseExpression()
    return BindingStmt {
        Name: name.Lexeme,
        Type: typ,
        Value: value,
    }
}

func (p *Parser) parseStatement() Stmt {
    switch p.current.Type {
    case TokenIdentifier:
        return p.parseBinding()

    default:
        p.error(fmt.Sprintf("unexpected %s", p.current.Lexeme))
        return nil
    }
}

func (p *Parser) Parse() *Program {
    program := &Program{}

    for !p.checkCurrent(TokenEOF) {
        if p.checkCurrent(TokenNewline) {
            p.advance()
            continue
        }

        program.Statements = append(program.Statements, p.parseStatement())
    }

    return program
}
