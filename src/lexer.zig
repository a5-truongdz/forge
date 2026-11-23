const std = @import("std");

const TokenType = enum {
    identifier,
    eof,

    colon,    // :
    equal,    // =
    comma,    // ,
    lt,       // <
    gt,       // >
    not,      // !
    minus,    // -
    multiply, // *
    divide,   // /
    power,    // ^
    hash,     // #
    modulo,   // %

    l_paren,  // ()
    r_paren,
    l_curly,  // {}
    r_curly,
    l_square, // []
    r_square,

    declare,     // :=
    eq,          // ==
    not_eq,      // !=
    le,          // <=
    ge,          // >=
    b_access,    // ->
    p_access,    // ::
    dereference, // >>
    address_of,  // <<

    kw_public,
    kw_private,
    kw_use,
    kw_class,
    kw_struct,
    kw_mut,
    kw_return,
    kw_true,
    kw_false,
    kw_for,
    kw_when,
    kw_if,
    kw_elif,
    kw_else,
    kw_forge,
    kw_break,
    kw_continue,
    kw_asm,
    kw_and,
    kw_or,

    integer,
    float,
    char,
    string
};

const keywords = std.StaticStringMap(TokenType).initComptime(.{
    .{ "public",   .kw_public },
    .{ "private",  .kw_private },
    .{ "use",      .kw_use },
    .{ "class",    .kw_class },
    .{ "struct",   .kw_struct },
    .{ "mut",      .kw_mut },
    .{ "return",   .kw_return },
    .{ "true",     .kw_true },
    .{ "false",    .kw_false },
    .{ "for",      .kw_for },
    .{ "when",     .kw_when },
    .{ "if",       .kw_if },
    .{ "elif",     .kw_elif },
    .{ "else",     .kw_else },
    .{ "forge",    .kw_forge },
    .{ "break",    .kw_break },
    .{ "continue", .kw_continue },
    .{ "asm",      .kw_asm },
    .{ "and",      .kw_and },
    .{ "or",       .kw_or }
});

const double_symbols = std.StaticStringMap(TokenType).initComptime(.{
    .{ ":=", .declare },
    .{ "==", .eq },
    .{ "!=", .not_eq },
    .{ "<=", .le },
    .{ ">=", .ge },
    .{ "->", .b_access },
    .{ "::", .p_access },
    .{ ">>", .dereference },
    .{ "<<", .address_of }
});

const TokenValue = union(enum) {
    none,
    identifier: []const u8,
    integer: i64,
    float: f64,
    string: []const u8,
    char: u8
};

const Loc = struct {
    file_path: []const u8,
    row: usize,
    column: usize,

    pub fn display(self: *const @This(), gpa: std.mem.Allocator) ![]u8 {
        return try gpa.print("{s}:{d}:{d}", .{ self.file_path, self.row, self.column });
    }
};

const Token = struct {
    typ: TokenType,
    value: TokenValue,
    loc: Loc,
};

const Diagnostic = struct {
    loc: Loc,
    message: []const u8,

    // remember to free()!!
    // const diag_display = lexResult.err.display(gpa);
    // defer gpa.free(diag_display);
    // ...
    pub fn display(self: *const @This(), gpa: std.mem.Allocator) ![]u8 {
        const loc_msg = try self.loc.display(gpa);
        defer gpa.free(loc_msg);
        return try gpa.print("{s}: {s}", .{ loc_msg, self.message });
    }
};

const LexResult = union(enum) {
    token: Token,
    err: Diagnostic
};

const EscapeResult = union(enum) {
    value: u8,
    err: Diagnostic
};

const Lexer = struct {
    const Self = @This();

    file_path: []const u8,
    source: []const u8,
    pos: usize = 0,
    bol: usize = 0,
    row: usize = 1,

    pub fn init(file_path: []const u8, source: []const u8) Self {
        return .{
            .file_path = file_path,
            .source = source,
        };
    }

    fn peek(self: *const Self) ?u8 {
        return if (self.pos < self.source.len) {
            self.source[self.pos];
        } else {
            null;
        };
    }

    fn advance(self: *Self) void {
        const ch = self.peek() orelse return;
        self.pos += 1;

        if (ch == '\n') {
            self.bol = self.pos;
            self.row += 1;
        }
    }

    // we are memory efficient uwu
    // fn loc(self: *const Self) *Loc {
        // return &.{
            // .file_path = self.file_path,
            // .row = self.row,
            // .column = self.pos - self.bol + 1,
        // };
    // }

    // no we cant be memory efficient anymore =(
    fn makeLoc(self: *const Self) Loc {
        return .{
            .file_path = self.file_path,
            .row = self.row,
            .column = self.pos - self.bol + 1,
        };
    }

    fn isIdentifierStart(ch: u8) bool {
        return std.ascii.isAlphabetic(ch) or ch == '_';
    }

    fn isIdentifierPart(ch: u8) bool {
        return std.ascii.isAlphanumeric(ch) or ch == '_';
    }

    fn makeToken(loc: Loc, typ: TokenType, val: TokenValue) LexResult {
        return .{
            .token = .{
                .loc = loc,
                .typ = typ,
                .value = val,
            }
        };
    }

    fn makeError(loc: Loc, msg: []const u8) LexResult {
        return .{
            .err = .{
                .loc = loc,
                .message = msg,
            }
        };
    }

    fn lexIdentifier(self: *Self, loc: Loc, start: usize) LexResult {
        while (self.peek()) |ch| {
            if (!isIdentifierPart(ch)) {
                break;
            }
            self.advance();
        }

        const text = self.source[start..self.pos];
        const typ = keywords.get(text) orelse .identifier;
        const val: TokenValue = switch (typ) {
            .identifier => .{ .identifier = text },
            else => .none
        };

        return makeToken(loc, typ, val);
    }

    fn lexEscape(self: *Self, loc: Loc) EscapeResult {
        self.advance();

        const ch = self.peek() orelse {
            return .{
                .err = .{
                    .loc = loc,
                    .message = "unterminated escape sequence",
                }
            };
        };
        self.advance();

        const val: u8 = switch (ch) {
            'n'  => '\n',
            'r'  => '\r',
            't'  => '\t',
            '"'  => '"',
            '0'  => 0,
            '\'' => '\'',
            '\\' => '\\',
            else => return .{
                .err = .{
                    .loc = loc,
                    .message = "invalid escape sequence",
                }
            }
        };

        return .{ .value = val };
    }

    fn lexNumber(self: *Self, gpa: std.mem.Allocator, loc: Loc, start: usize) !LexResult {
        var has_dot = false;
        var has_e = false;

        while (self.peek()) |ch| {
            if (std.ascii.isDigit(ch)) {
                self.advance();
            } else if (ch == '.' and !has_dot) {
                if (has_e) {
                    // e first, then dot is invalid
                    return makeError(loc, "decimal point cannot appear after exponent");
                }

                has_dot = true;
                self.advance();

                const c = self.peek() orelse {
                    return makeError(loc, "expected digit after '.'");
                };

                if (!std.ascii.isDigit(c)) {
                    return makeError(loc, "expected digit after '.'");
                }
                self.advance();
            } else if (ch == 'e' and !has_e) {
                has_e = true;
                self.advance();

                const c = self.peek() orelse {
                    return makeError(loc, "expected digit after exponent");
                };

                if (c == '+' or c == '-') {
                    self.advance();
                }

                const c2 = self.peek() orelse {
                    return makeError(loc, "expected digit after exponent");
                };

                if (!std.ascii.isDigit(c2)) {
                    return makeError(loc, "expected digit after exponent");
                }
                self.advance();
            } else {
                break;
            }
        }

        if (self.peek()) |ch| {
            switch (ch) {
                '.' => return makeError(loc, "multiple decimal points in number"),
                'e' => return makeError(loc, "multiple exponent symbol in number"),
                else => {}
            }
        }

        const text = self.source[start..self.pos];

        // i think using arena allocator is better here
        // since errors are printed at the very last phase of compilaion
        // then exits right after
        // it will auto clean up ig
        // but imma keep using gpa
        // more suffer when .free()-ing yk
        if (has_dot || has_e) {
            // parseFloat() do parse exponent and +/- natively!!!
            const val = std.fmt.parseFloat(f64, text) catch |err| {
                return makeError(loc, try gpa.print("parsing floating-point number failed: {}", .{ err }));
            };
            return makeToken(loc, .float, .{ .float = val });
        }

        const val = std.fmt.parseInt(i64, text, 10) catch |err| {
            return makeError(loc, try gpa.print("parsing integer failed: {}", .{ err }));
        };
        return makeToken(loc, .integer, .{ .integer = val }); 
    }

    fn lexString(self: *Self, gpa: std.mem.Allocator, loc: Loc) !LexResult {
        self.advance();

        var str_buf: std.ArrayList(u8) = .empty;
        errdefer str_buf.deinit(gpa);

        while (self.peek()) |ch| {
            switch (ch) {
                // who gives a fuck abt \r
                // we are linux users
                '\n' => return makeError(loc, "unterminated string"),
                '"'  => {
                    self.advance();
                    const value = try str_buf.toOwnedSlice(gpa);
                    return makeToken(loc, .string, .{ .string = value });
                },

                // escape stuff idk
                '\\' => {
                    switch (self.lexEscape(self.makeLoc())) {
                        .value => |val| try str_buf.append(gpa, val),
                        .err   => |err| return .{ .err = err }
                    }
                },

                else => {
                    self.advance();
                    try str_buf.append(gpa, ch);
                }
            }
        }

        return makeError(loc, "unterminated string");
    }

    fn lexChar(self: *Self, loc: Loc) LexResult {
        self.advance();

        const ch = self.peek() orelse {
            return makeError(loc, "unterminated character");
        };

        const value: u8 = if (ch == '\\') {
            switch (self.lexEscape(self.makeLoc())) {
                .value => |val| val,
                .err   => |err| return .{ .err = err }
            }
        } else {
            switch (ch) {
                '\n' => return makeError(loc, "unterminated character"),
                '\'' => return makeError(loc, "empty character literal"),
                else => blk: {
                    self.advance();
                    break :blk ch;
                }
            }
        };

        if (self.peek() != '\'') {
            return makeError(loc, "character literal must contain one character");
        }
        self.advance();

        return makeToken(loc, .char, .{ .char = value });
    }

    fn lexSymbol(self: *Self, loc: Loc) LexResult {
        // double char
        if (self.pos + 1 < self.source.len) {
            const symbol = self.source[self.pos..self.pos + 2];
            if (double_symbols.get(symbol)) |typ| {
                self.advance();
                self.advance();
                return makeToken(loc, typ, .none);
            }
        }

        // single char
        const ch = self.peek().?;
        self.advance();

        const typ: TokenType = switch (ch) {
            ':' => .colon,
            '=' => .equal,
            ',' => .comma,
            '<' => .lt,
            '>' => .gt,
            '!' => .not,
            '-' => .minus,
            '*' => .multiply,
            '/' => .divide,
            '^' => .power,
            '#' => .hash,
            '%' => .modulo,

            '(' => .l_paren,
            ')' => .r_paren,
            '{' => .l_curly,
            '}' => .r_curly,
            '[' => .l_square,
            ']' => .r_square,

            else => return makeError(loc, "unexpected character")
        };

        return makeToken(loc, typ, .none);
    }

    pub fn nextToken(self: *Self, gpa: std.mem.Allocator) !LexResult {
        // ppl keeps telling me to put this on a function
        // but why
        // useless af
        // called only once too
        while (self.peek()) |ch| {
            if (!std.ascii.isWhitespace(ch)) {
                break;
            }
            self.advance();
        }

        const start = self.pos;
        const loc = self.makeLoc();
        const ch = self.peek() orelse return makeToken(loc, .eof, .none);

        if (isIdentifierStart(ch)) {
            return self.lexIdentifier(loc, start);
        } else if (std.ascii.isDigit(ch)) {
            return self.lexNumber(gpa, loc, start);
        } else if (ch == '"') {
            return self.lexString(gpa, loc);
        } else if (ch == '\'') {
            return self.lexChar(loc);
        } else {
            return self.lexSymbol(loc);
        }
    }
};
