# Needle

## AST

### Statements

- Block `{}`
- Expression `;`
- Declaration `var` / `const`
- Assignment `=`
- If `if (); else;`
- While `while ();`
- Do `do; while();` **TODO**
- Say `say;`
- Import `import . "module";` **TODO**
- Export `export value;` **TODO**

### Expressions

- Infix `a + b`
- Prefix `not a`

### Literals

- null `null`
- boolean `true` / `false`
- number `12.3`
- string `"Hello, World!"`
- function `fun(a, b) { return a + b; }` **TODO**

### Operators

#### Infix

- plus `+`
- minus `-`
- star `*`
- slash `/`
- or `or`
- and `and`
- eq `==`
- ne `!=`
- lt `<`
- le `<=`
- gt `>`
- ge `>=`

#### Prefix

- not `not`
- plus `+`
- minus `-`


## Code

```typescript
var a;
const c = "true";

a = not c;

if (a < 2) {
    1 + 2;
} else if (a >= 2) 2 + 4;
else 2 + 2;

while (a) 2 + 2;
while (b) {
    2 + 2;
}

say 2 + 2;
```
