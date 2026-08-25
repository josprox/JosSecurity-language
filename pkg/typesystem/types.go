// Package typesystem is the canonical source of truth for Joss type names,
// inference and assignment compatibility.
package typesystem

import (
	"math"
	"strconv"
	"strings"
)

type Kind string

const (
	Unknown Kind = "unknown"
	Mixed   Kind = "mixed"
	Null    Kind = "null"
	Int     Kind = "int"
	Float   Kind = "float"
	String  Kind = "string"
	Bool    Kind = "bool"
	Array   Kind = "array"
	Map     Kind = "map"
	Object  Kind = "object"
	Channel Kind = "channel"
	Class   Kind = "class"
	Union   Kind = "union"
)

// Type represents a semantic type. Name is populated for class types.
type Type struct {
	Kind Kind
	Name string
}

func (t Type) String() string {
	if (t.Kind == Class || t.Kind == Union) && t.Name != "" {
		return t.Name
	}
	if t.Kind == "" {
		return string(Unknown)
	}
	return string(t.Kind)
}

func (t Type) IsKnown() bool {
	return t.Kind != "" && t.Kind != Unknown && t.Kind != Null && t.Kind != Mixed
}
func (t Type) IsDynamic() bool { return t.Kind == Mixed }

// Members returns the normalized alternatives of a union. Non-union types
// return themselves as a single member. Type intentionally remains comparable;
// the canonical union spelling is stored in Name rather than a slice.
func (t Type) Members() []Type {
	if t.Kind != Union {
		return []Type{t}
	}
	parts := strings.Split(t.Name, "|")
	result := make([]Type, 0, len(parts))
	for _, part := range parts {
		result = append(result, Parse(part))
	}
	return result
}

// Parse returns the canonical meaning of a source-level type name.
func Parse(name string) Type {
	name = strings.TrimSpace(name)
	if strings.HasSuffix(name, "?") {
		name = strings.TrimSpace(strings.TrimSuffix(name, "?")) + "|null"
	}
	if strings.Contains(name, "|") {
		seen := make(map[string]bool)
		members := make([]string, 0)
		for _, part := range strings.Split(name, "|") {
			member := Parse(part)
			for _, flattened := range member.Members() {
				canonical := flattened.String()
				if canonical == "" || seen[canonical] {
					continue
				}
				seen[canonical] = true
				members = append(members, canonical)
			}
		}
		if len(members) == 1 {
			return Parse(members[0])
		}
		return Type{Kind: Union, Name: strings.Join(members, "|")}
	}
	switch strings.ToLower(name) {
	case "", "unknown", "var":
		return Type{Kind: Unknown}
	case "mixed":
		return Type{Kind: Mixed}
	case "null", "nil":
		return Type{Kind: Null}
	case "int":
		return Type{Kind: Int}
	case "float":
		return Type{Kind: Float}
	case "string":
		return Type{Kind: String}
	case "bool":
		return Type{Kind: Bool}
	case "array":
		return Type{Kind: Array}
	case "map":
		return Type{Kind: Map}
	case "object":
		return Type{Kind: Object}
	case "channel":
		return Type{Kind: Channel}
	default:
		return Type{Kind: Class, Name: name}
	}
}

// Assignable reports whether a value of source type can be assigned to a
// destination. Unknown is deliberately non-accusatory: lack of information is
// not evidence of invalid user code.
func Assignable(destination, source Type) bool {
	if destination.Kind == Mixed || source.Kind == Mixed || destination.Kind == Unknown || source.Kind == Unknown {
		return true
	}
	if source.Kind == Union {
		for _, sourceMember := range source.Members() {
			if !Assignable(destination, sourceMember) {
				return false
			}
		}
		return true
	}
	if destination.Kind == Union {
		for _, destinationMember := range destination.Members() {
			if Assignable(destinationMember, source) {
				return true
			}
		}
		return false
	}
	if destination.Kind == source.Kind {
		return destination.Kind != Class || destination.Name == source.Name
	}
	// Integer values are losslessly accepted by float variables, matching the
	// runtime's existing numeric compatibility rule.
	if destination.Kind == Float && source.Kind == Int {
		return true
	}
	if destination.Kind == Object && source.Kind == Class {
		return true
	}
	return false
}

// MergeInference keeps the first concrete inferred type. Null/unknown
// initializers postpone inference until a concrete value is assigned.
func MergeInference(current, candidate Type) Type {
	if current.Kind == Mixed {
		return current
	}
	if !current.IsKnown() && candidate.IsKnown() {
		return candidate
	}
	if current.Kind == "" {
		return Type{Kind: Unknown}
	}
	return current
}

// CoerceString applies the language's explicit typed-input conversion policy.
// It is shared by the runtime and semantic analyzer so a literal accepted by
// one phase cannot be rejected by the other.
func CoerceString(destination Type, value string) (interface{}, bool) {
	if destination.Kind == Union {
		for _, member := range destination.Members() {
			if coerced, ok := CoerceString(member, value); ok {
				return coerced, true
			}
		}
		return value, false
	}
	value = strings.TrimSpace(value)
	switch destination.Kind {
	case Int:
		if integer, err := strconv.ParseInt(value, 10, 64); err == nil {
			return integer, true
		}
		if number, err := strconv.ParseFloat(value, 64); err == nil &&
			math.Trunc(number) == number && number >= float64(math.MinInt64) && number < float64(math.MaxInt64) {
			return int64(number), true
		}
	case Float:
		if number, err := strconv.ParseFloat(value, 64); err == nil {
			return number, true
		}
	case Bool:
		switch strings.ToLower(value) {
		case "true", "1", "yes":
			return true, true
		case "false", "0", "no", "":
			return false, true
		}
	}
	return value, false
}

// SourceTypeNames returns the supported source-level type spellings used by
// editor/tooling catalog generation.
func SourceTypeNames() []string {
	return []string{"array", "bool", "channel", "float", "int", "map", "mixed", "object", "string", "var"}
}
