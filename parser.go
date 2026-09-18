package main

type Parser struct {
    lexer *Lexer
}

func NewParser(lexer *Lexer) *Parser {
    return &Parser {
        lexer: lexer,
    }
}

func (p *Parser) __suppress() {}
