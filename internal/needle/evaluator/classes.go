package evaluator

import "strconv"

const (
	CLASS_NUMBER = "Number"
	CLASS_STRING = "String"
	CLASS_ARRAY  = "Array"
	CLASS_TABLE  = "Table"
)

func newNumberClass() *Class {
	return &Class{
		Public: map[string]*Function{
			"to_string": {
				FType: F_NATIVE,
				Native: coverNative(func(e *Evaluator, this Value, args ...Value) Value {
					num := this.(*Number)
					return &String{Value: strconv.FormatFloat(num.Value, 'g', -1, 64)}
				}, 0),
			},
			"to_boolean": {
				FType: F_NATIVE,
				Native: coverNative(func(e *Evaluator, this Value, args ...Value) Value {
					num := this.(*Number)
					return &Boolean{Value: num.Value != 0}
				}, 0),
			},
		},
	}
}

func newStringClass() *Class {
	return &Class{
		Public: map[string]*Function{
			"reverse": {
				FType: F_NATIVE,
				Native: coverNative(func(e *Evaluator, this Value, args ...Value) Value {
					str := this.(*String)
					rev := []rune(str.Value)
					for i := 0; i < len(rev)/2; i++ {
						alt := len(rev) - i - 1
						rev[i], rev[alt] = rev[alt], rev[i]
					}
					return &String{Value: string(rev)}
				}, 0),
			},
			"to_upper_case": {
				FType: F_NATIVE,
				Native: coverNative(func(e *Evaluator, this Value, args ...Value) Value {
					str := this.(*String)
					up := []rune(str.Value)
					for i := range len(up) {
						if 'a' <= up[i] && up[i] <= 'z' {
							up[i] += 'A' - 'a'
						}
					}
					return &String{Value: string(up)}
				}, 0),
			},
			"to_lower_case": {
				FType: F_NATIVE,
				Native: coverNative(func(e *Evaluator, this Value, args ...Value) Value {
					str := this.(*String)
					low := []rune(str.Value)
					for i := range len(low) {
						if 'A' <= low[i] && low[i] <= 'Z' {
							low[i] -= 'A' - 'a'
						}
					}
					return &String{Value: string(low)}
				}, 0),
			},
		},
	}
}

func newArrayClass() *Class {
	return &Class{
		Public: map[string]*Function{
			"push": {
				FType: F_NATIVE,
				Native: coverNative(func(e *Evaluator, this Value, args ...Value) Value {
					arr := this.(*Array)
					arr.Elements = append(arr.Elements, args...)
					return e.env.globals.Null
				}, 1),
			},
			"pop": {
				FType: F_NATIVE,
				Native: coverNative(func(e *Evaluator, this Value, args ...Value) Value {
					arr := this.(*Array)
					if len(arr.Elements) == 0 {
						e.ThrowException("array is empty")
					}
					elem := arr.Elements[len(arr.Elements)-1]
					arr.Elements = arr.Elements[:len(arr.Elements)-1]
					return elem
				}, 0),
			},
			"length": {
				FType: F_NATIVE,
				Native: coverNative(func(e *Evaluator, this Value, args ...Value) Value {
					arr := this.(*Array)
					return &Number{Value: float64(len(arr.Elements))}
				}, 0),
			},
		},
	}
}

func newTableClass() *Class {
	return &Class{
		Public: map[string]*Function{
			"size": {
				FType: F_NATIVE,
				Native: coverNative(func(e *Evaluator, this Value, args ...Value) Value {
					tbl := this.(*Table)
					return &Number{Value: float64(tbl.Pairs.Size())}
				}, 0),
			},
			"delete": {
				FType: F_NATIVE,
				Native: coverNative(func(e *Evaluator, this Value, args ...Value) Value {
					tbl := this.(*Table)
					exist, err := tbl.Pairs.Delete(args[0])
					if err != nil {
						e.ThrowException("%s", err.Error())
					}
					return &Boolean{Value: exist}
				}, 1),
			},
		},
	}
}

func CreateBaseClasses() map[string]*Class {
	classes := map[string]*Class{}
	classes[CLASS_NUMBER] = newNumberClass()
	classes[CLASS_STRING] = newStringClass()
	classes[CLASS_ARRAY] = newArrayClass()
	classes[CLASS_TABLE] = newTableClass()
	return classes
}
