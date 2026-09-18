package main

import "fmt"

type TokenEnum int

const (
    TokenEOF TokenEnum = iota
    TokenNewline

    // Symbols
    TokenColon    // :
    TokenEqual    // =
    TokenComma    // ,
    TokenLT       // <
    TokenGT       // >
    TokenNot      // !
    TokenPipe     // |
    TokenPlus     // +
    TokenMinus    // -
    TokenMultiply // *
    TokenDivide   // /
    TokenPower    // ^
    TokenDot      // .
    TokenHash     // #
    TokenModulo   // %

    // Brackets
    TokenLParen    // ()
    TokenRParen
    TokenLCurly    // {}
    TokenRCurly
    TokenLSquare   // []
    TokenRSquare

    // Double-char symbols
    TokenDeclare     // :=
    TokenEQ          // ==
    TokenNotEQ       // !=
    TokenLE          // <=
    TokenGE          // >=
    TokenBAccess     // ->
    TokenPAccess     // ::
    TokenDoubleDots  // ..
    TokenDereference // >>
    TokenAddressOf   // <<
    TokenAnd         // &&
    TokenOr          // ||

    // Keywords
    TokenPublic
    TokenPrivate
    TokenUse
    TokenClass
    TokenStruct
    TokenMut
    TokenReturn
    TokenInt
    TokenFloat
    TokenBool
    TokenChar
    TokenString
    TokenTrue
    TokenFalse
    TokenFor
    TokenWhen
    TokenIf
    TokenElif
    TokenElse
    TokenForge
    TokenBreak
    TokenContinue
    TokenType
    TokenAsm

    // Literals
    TokenIdentifier
    TokenLInt
    TokenLFloat
    TokenLChar
    TokenLString
)

var keywords = map[string]TokenEnum {
    "public":   TokenPublic,
    "private":  TokenPrivate,
    "use":      TokenUse,
    "class":    TokenClass,
    "struct":   TokenStruct,
    "mut":      TokenMut,
    "return":   TokenReturn,
    "int":      TokenInt,
    "float":    TokenFloat,
    "bool":     TokenBool,
    "char":     TokenChar,
    "string":   TokenString,
    "true":     TokenTrue,
    "false":    TokenFalse,
    "for":      TokenFor,
    "when":     TokenWhen,
    "if":       TokenIf,
    "elif":     TokenElif,
    "else":     TokenElse,
    "forge":    TokenForge,
    "break":    TokenBreak,
    "continue": TokenContinue,
    "type":     TokenType,
    "asm":      TokenAsm,
}

var doubleSymbols = map[string]TokenEnum{
    ":=": TokenDeclare,
    "==": TokenEQ,
    "!=": TokenNotEQ,
    "<=": TokenLE,
    ">=": TokenGE,
    "->": TokenBAccess,
    "::": TokenPAccess,
    "..": TokenDoubleDots,
    ">>": TokenDereference,
    "<<": TokenAddressOf,
    "&&": TokenAnd,
    "||": TokenOr,
}

var singleSymbols = map[byte]TokenEnum{
    ':': TokenColon,
    '=': TokenEqual,
    ',': TokenComma,
    '<': TokenLT,
    '>': TokenGT,
    '!': TokenNot,
    '|': TokenPipe,
    '+': TokenPlus,
    '-': TokenMinus,
    '*': TokenMultiply,
    '/': TokenDivide,
    '^': TokenPower,
    '.': TokenDot,
    '#': TokenHash,
    '%': TokenModulo,

    '(': TokenLParen,
    ')': TokenRParen,
    '{': TokenLCurly,
    '}': TokenRCurly,
    '[': TokenLSquare,
    ']': TokenRSquare,
}

type Token struct {
    Type TokenEnum
    Lexeme string
    Line int
    Column int
}

type Lexer struct {
    source string
    filename string
    pos int
    line int
    column int
}

type LexerError struct {
    Filename string
    Line     int
    Column   int
    Message  string
}

func (e LexerError) Error() string {
    return fmt.Sprintf(
        "%s:%d:%d: %s",
        e.Filename,
        e.Line,
        e.Column,
        e.Message,
    )
}

func NewLexer(filename, src string) *Lexer {
    return &Lexer {
        source: src,
        filename: filename,
        pos: 0,
        line: 1,
        column: 1,
    }
}

func (l *Lexer) eof() bool {
    return l.pos >= len(l.source)
}

func (l *Lexer) peek() byte {
    if l.eof() {
        return 0
    }

    return l.source[l.pos]
}

func (l *Lexer) advance() byte {
    c := l.peek()
    if c == 0 {
        return 0
    }

    l.pos++

    if c == '\n' {
        l.line++
        l.column = 1
    } else {
        l.column++
    }

    return c
}

func (l *Lexer) throwError(line, col int, message string) {
    panic(LexerError{
        Filename: l.filename,
        Line:     line,
        Column:   col,
        Message:  message,
    })
}


func (l *Lexer) makeToken(
    typ TokenEnum,
    start int,
    line,
    column int,
) Token {
    return Token{
        Type:   typ,
        Lexeme: l.source[start:l.pos],
        Line:   line,
        Column: column,
    }
}

func (l *Lexer) lexIdentifier(start, line, col int) Token {
    for isIdentifierPart(l.peek()) {
        l.advance()
    }

    typ, ok := keywords[l.source[start:l.pos]]
    if ok {
        return l.makeToken(typ, start, line, col)
    } else {
        return l.makeToken(TokenIdentifier, start, line, col)
    }
}

func (l *Lexer) lexNumber(start, line, col int) Token {
    hasDot := false
    hasE := false

    for isDigit(l.peek()) {
        l.advance()
    }

    if l.peek() == '.' {
        hasDot = true
        l.advance()

        if !isDigit(l.peek()) {
            l.throwError(line, col, "expected digit after '.'")
        }

        for isDigit(l.peek()) {
            l.advance()
        }
    }

    if l.peek() == 'e' {
        hasE = true
        l.advance()

        if l.peek() == '+' || l.peek() == '-' {
            l.advance()
        }

        if !isDigit(l.peek()) {
            l.throwError(line, col, "expected digit in exponent")
        }

        for isDigit(l.peek()) {
            l.advance()
        }
    }

    if l.peek() == '.' {
        l.throwError(line, col, "multiple decimal points in number")
    }

    if hasDot || hasE {
        return l.makeToken(TokenLFloat, start, line, col)
    }

    return l.makeToken(TokenLInt, start, line, col)
}

func (l *Lexer) lexString(start, line, col int) Token {
    l.advance()

    for {
        c := l.peek()

        switch c {
        case 0, '\n':
            l.throwError(line, col, "unterminated string")

        case '"':
            l.advance()
            return l.makeToken(TokenLString, start, line, col)

        case '\\':
            l.advance()

            switch l.peek() {
            case 'n', 'r', '"', '\'', '\\':
                l.advance()

            default:
                l.throwError(l.line, l.column, "invalid escape sequence")
            }

        default:
            l.advance()
        }
    }
}

func (l *Lexer) lexChar(start, line, col int) Token {
    l.advance()

    c := l.peek()

    if c == 0 || c == '\n' {
        l.throwError(line, col, "unterminated character")
    }

    if c == '\\' {
        l.advance()

        switch l.peek() {
        case 'n', 'r', '\'', '"', '\\':
            l.advance()

        default:
            l.throwError(l.line, l.column, "invalid escape sequence")
        }
    } else {
        l.advance()
    }

    if l.peek() != '\'' {
        l.throwError(line, col, "character literal must contain one character")
    }

    l.advance()
    return l.makeToken(TokenLChar, start, line, col)
}

func (l *Lexer) lexSymbol(start, line, col int) Token {
    if l.pos + 1 < len(l.source) {
        symbol := l.source[l.pos : l.pos+2]

        if typ, ok := doubleSymbols[symbol]; ok {
            l.advance()
            l.advance()
            return l.makeToken(typ, start, line, col)
        }
    }

    c := l.peek()
    if typ, ok := singleSymbols[c]; ok {
        l.advance()
        return l.makeToken(typ, start, line, col)
    }

    l.throwError(line, col, "unexpected character")
    return Token{} // unreachable
}

func (l *Lexer) skipSingleLineComment() {
    for !l.eof() && l.peek() != '\n' {
        l.advance()
    }

    // dont consume the '\n' itself here
    // so we have TokenNewline
}

func (l *Lexer) skipMultiLineComment() {
    line, col := l.line, l.column

    for {
        if l.eof() {
            l.throwError(line, col, "unterminated multi-line comment")
        }

        if l.pos+3 <= len(l.source) && l.source[l.pos:l.pos+3] == "-->" {
            l.advance()    // '-'
            l.advance()    // '-'
            l.advance()    // '>'
            return
        }

        l.advance()
    }
}

func (l *Lexer) skipComment() {
    l.advance()    // '<'
    l.advance()    // '-'
    l.advance()    // '-'

    if l.peek() == '\n' {
        l.skipMultiLineComment()
        return
    }

    l.skipSingleLineComment()
}

func (l *Lexer) Next() Token {
    for isWhitespace(l.peek()) {
        l.advance()
    }

    start := l.pos
    line := l.line
    col := l.column
    c := l.peek()

    switch {
    case c == 0:
        return l.makeToken(TokenEOF, start, line, col)

    case isIdentifierStart(c):
        return l.lexIdentifier(start, line, col)

    case isDigit(c):
        return l.lexNumber(start, line, col)

    case c == '"':
        return l.lexString(start, line, col)

    case c == '\'':
        return l.lexChar(start, line, col)

    case c == '\n':
        l.advance()
        return l.makeToken(TokenNewline, start, line, col)

    case c == '<' && l.pos+3 <= len(l.source) && l.source[l.pos:l.pos+3] == "<--":
        l.skipComment()
        return l.Next()

    default:
        return l.lexSymbol(start, line, col)
    }
}
