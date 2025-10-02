# Needle

## AST

### Statements

- Block `{}`
- Expression `;`
- Declaration `var` / `const` / `fun`
- Assignment `=`
- If `if (); else;`
- For `for (;;);` **TODO**
- While `while ();`
- Do `do; while();` **TODO**
- Say `say;`
- Import `import . "module";` **TODO**
- Export `export value;` **TODO**
- Try `try; catch(); finally;`
- Throw `throw;`
- Return `return;`
- Break `break;`
- Continue `continue;`

### Expressions

- Infix `a + b`
- Prefix `not a`
- Call `()`

### Literals

- null `null`
- boolean `true` / `false`
- number `12.3`
- string `"Hello, World!"`
- function `fun(a, b) { return a + b; }`

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
- is `===`
- isnt `!==`
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

class Document {
    var name: String
    constructor new(name) {
        this.name= name;
    }

    get .(attr: String): any {
        return "Hello";
    }
    get name(): any {
        return "Lin";
    }
    get [](index: any): any {
        if (type_of(index) != "number") throw Error("not number index");
        return this.list[index];
    }
    get [:](start: any, end: any): any {
        if (type_of(start) != "number" or
            type_of(end) != "number")
            throw Error("not number index");
        return this.list[start:end];
    }

    set .(attr: String, value: any) {
        this.dict[attr] = value;
    }
    set name(name: any) {
        this.name = name;
    }
    set [](index: any, value: any) {
        if (type_of(index) != "number") throw Error("not number index");
        return this.list[index] = value;
    }
    set [:](start: any, end: any, value: any): any {
        if (type_of(start) != "number" or
            type_of(end) != "number")
            throw Error("not number index");
        return this.list[start:end] = value;
    }

    infix plus(other: any): any {  // document plus "haha";
        return this.name + other;
    }
    infix +(other: any): any {
        return this.name + other;
    }
    infix -(other: any): any {
        return this.name - other;
    }
    infix /(other: any): any {
        return this.name / other;
    }
    infix *(other: any): any {
        return this.name * other;
    }
    infix ==(other: any): any {
        return this.name == other;
    }
    infix <(other: any): any {
        return this.name < other;
    }
    infix <=(other: any): any {
        return this.name <= other;
    }
}

class {
    constructor new() {

    }
    var a;
    var b = "Hello!";
    fun hello() {
        say b;
    }
}.new()

const arr = array{1, 2, 3, 4, [11] = null};
arr[3]  // 4
const hi = 123;
const m = map{
    ["hello"] = "hello",
    [hi] = hi,
};

```
