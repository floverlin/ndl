# Needle

the n33d<= programming language

## Syntax

```
program         -> declaration* EOF ;
```

### Declarations

```
declaration     -> varDecl
                 | statement ;
varDecl         -> "var" IDENTIFIER ( "=" expression )? ";" ;
```

### Statements

```
statement       -> exprStmt
                 | assignStmt
                 | whileStmt
                 | ifStmt
                 | tryStmt
                 | returnStmt
                 | breakStmt
                 | continueStmt
                 | sayStmt
                 | block ;
exprStmt        -> expression ";" ;
assignStmt      -> ( propExpr | indexExpr | sliceExpr | IDENTIFIER )
                 "=" expression ";" ;
whileStmt       -> "while" "(" expression ")" statement ;
ifStmt          -> "if" "(" expression ")" statement
                 ( "else" statement )? ;
tryStmt         -> "try" statement
                 ( "catch" "(" IDENTIFIER ")" statement )?
                 ( "finally" statement )? ;
returnStmt      -> "return" ( expression )? ";" ;
breakStmt       -> "break" ";" ;
continueStmt    -> "continue" ";" ;
sayStmt         -> "say" expression ";" ;
block           -> "{" declaration* "}" ;
```

### Expressions

- left associativity

```
expression      -> prefixExpr
                 | infixExpr
                 | indexExpr
                 | sliceExpr
                 | callExpr
                 | propExpr
                 | group
                 | literal ;
prefixExpr      -> prefix_operator expression ;
infixExpr       -> expression infix_operator expression ;
indexExpr       -> expression "[" expression "]" ;
sliceExpr       -> expression "[" expression ":" expression "]" ;
callExpr        -> expression "(" arguments? ")" ;
propExpr        -> expression "." IDENTIFIER ;
group           -> "(" expression ")" ;
literal         -> "true" | "false" | "null" | "this"
                 | NUMBER | STRING | IDENTIFIER
                 | FUNCTION | CLASS | ARRAY | MAP
```

### Operators

```
prefix_operator -> ( "+" | "-" | "!" ) ;
infix_operator  -> logic_or
                 | logic_and
                 | equality
                 | comparision
                 | term
                 | factor ;
logic_or        -> "or" ;
logic_and       -> "and" ;
equality        -> "==" | "!=" | "===" | "!==" ;
comparision     -> "<" | ">" | "<=" | ">=" ;
term            -> "+" | "-" ;
factor          -> "*" | "/" ;
```

### Literals

```
NUMBER          -> DIGIT+ ( "." DIGIT+ )? ;
STRING          -> "\"" <any char except "\"">* "\"" ;
IDENTIFIER      -> ALPHA ( ALPHA | DIGIT )*
                 | "`" ALPHA ( ALPHA | DIGIT )* "`" ;
ALPHA           -> "a" ... "z" | "A" ... "Z" | "_" ;
DIGIT           -> "0" ... "9" ;
FUN             -> "fun" function ;
CLASS           -> "class" class ;
ARRAY           -> "array" array ;
MAP             -> "map" map ;
```

### Utility
```
function        -> "(" parameters? ")" block ;
class           -> "{" class_decl* "}" ;
array           -> "{" array_decl? "}" ;
map             -> "{" map_decl? "}" ;
arguments       -> expression ( "," expression )? ","? ;
parameters      -> IDENTIFIER ( "," IDENTIFIER )? ","? ;
class_decl      -> "constructor" IDENTIFIER function
                 | "public" IDENTIFIER function
                 | "private" IDENTIFIER function
                 | "get" ( IDENTIFIER | "." | "[]" | "[:]" ) function
                 | "set" ( IDENTIFIER | "." | "[]" | "[:]" ) function
                 | "infix" (IDENTIFIER | term | factor | "==" | "<" | "<=" )
                 function
                 | varDecl ;
array_decl      -> ( expression | "[" expresion "]" "=" expression )
                 ( "," expression  | "[" expresion "]" "=" expression )* ","? ;
map_dacl        -> "[" expresion "]" "=" expression
                 ( "," "[" expresion "]" "=" expression )? ","? ;
```

### Precedence

```
-HIGHEST-
    group
    call & slice & index & prop
    factor
    term
    comparision
    equality
    logic_and
    logic_or
    assign
-LOWER-
```