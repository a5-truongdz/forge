package main

import "os"

func main() {
    filename := "hello.forge"
    file, err := os.ReadFile(filename)
    if err != nil {
        panic("yo file broken gng")
    }

    lexer := NewLexer(filename, string(file))
    parser := NewParser(lexer)
    parser.Parse()

    // executor := NewExecutor(ast)    // or compiler if i want it to be asm
    // executor.Run()
}
