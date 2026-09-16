package main

import (
    "os"
    "fmt"
)

func main() {
    filename := "hello.forge"
    file, err := os.ReadFile(filename)
    if err != nil {
        panic("yo file broken gng")
    }

    lexer := NewLexer(filename, string(file))

    for {
        tok := lexer.Next()

        fmt.Printf("%-15s %-12q %d:%d\n",
            tok.Type,
            tok.Lexeme,
            tok.Line,
            tok.Column,
        )

        if tok.Type == TokenEOF {
            break
        }
    }

    // parser := NewParser(lexer)

    // ast := parser.Parse()
    // executor := NewExecutor(ast)    // or compiler if i want it to be asm
    // executor.Run()
}
