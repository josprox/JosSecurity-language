package vm

import "fmt"

type ValueKind uint8

const (
	ValNull ValueKind = iota
	ValBool
	ValInt
	ValFloat
	ValString
)

type Value struct {
	Kind    ValueKind
	Integer int64
	Float   float64
	Boolean bool
	Str     string
}

func NullVal() Value {
	return Value{Kind: ValNull}
}

func BoolVal(b bool) Value {
	return Value{Kind: ValBool, Boolean: b}
}

func IntVal(i int64) Value {
	return Value{Kind: ValInt, Integer: i}
}

func FloatVal(f float64) Value {
	return Value{Kind: ValFloat, Float: f}
}

func StringVal(s string) Value {
	return Value{Kind: ValString, Str: s}
}

func (v Value) String() string {
	switch v.Kind {
	case ValNull:
		return "null"
	case ValBool:
		if v.Boolean {
			return "true"
		}
		return "false"
	case ValInt:
		return fmt.Sprintf("%d", v.Integer)
	case ValFloat:
		return fmt.Sprintf("%g", v.Float)
	case ValString:
		return v.Str
	default:
		return "<unknown>"
	}
}

func (v Value) IsTruthy() bool {
	switch v.Kind {
	case ValNull:
		return false
	case ValBool:
		return v.Boolean
	case ValInt:
		return v.Integer != 0
	case ValFloat:
		return v.Float != 0.0
	case ValString:
		return v.Str != ""
	default:
		return false
	}
}
