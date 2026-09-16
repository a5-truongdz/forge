package main

func isDigit(c byte) bool {
    return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
    return (c >= 'a' && c <= 'z') ||
           (c >= 'A' && c <= 'Z')
}

func isIdentifierStart(c byte) bool {
    return isLetter(c) || c == '_'
}

func isIdentifierPart(c byte) bool {
    return isIdentifierStart(c) || isDigit(c)
}

func isWhitespace(c byte) bool {
    return c == ' '  ||
           c == '\r' ||
           c == '\t'
}
