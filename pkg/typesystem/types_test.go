package typesystem

import "testing"

func TestCanonicalNamesAndCompatibility(t *testing.T) {
	for _, removedAlias := range []string{"integer", "double", "boolean", "dynamic", "any", "list"} {
		if got := Parse(removedAlias); got.Kind != Class {
			t.Fatalf("removed alias %q resolved as %s, want unresolved class type", removedAlias, got.String())
		}
	}
	if !Assignable(Type{Kind: Float}, Type{Kind: Int}) {
		t.Fatal("int should be assignable to float")
	}
	if Assignable(Type{Kind: Int}, Type{Kind: String}) {
		t.Fatal("string must not be assignable to int")
	}
	if !Assignable(Type{Kind: Mixed}, Type{Kind: String}) {
		t.Fatal("mixed is the explicit dynamic escape hatch")
	}
}

func TestInferenceWaitsForConcreteValue(t *testing.T) {
	unknown := Type{Kind: Unknown}
	if got := MergeInference(unknown, Type{Kind: Null}); got.Kind != Unknown {
		t.Fatalf("null should postpone inference, got %s", got.String())
	}
	if got := MergeInference(unknown, Type{Kind: Int}); got.Kind != Int {
		t.Fatalf("int should establish inference, got %s", got.String())
	}
}

func TestTypedStringCoercionIsLossless(t *testing.T) {
	if value, ok := CoerceString(Type{Kind: Int}, "20"); !ok || value != int64(20) {
		t.Fatalf("integer coercion = %v, %v", value, ok)
	}
	if _, ok := CoerceString(Type{Kind: Int}, "20.5"); ok {
		t.Fatal("fractional input must not be truncated into int")
	}
	if _, ok := CoerceString(Type{Kind: Int}, "twenty"); ok {
		t.Fatal("non-numeric input must not coerce to int")
	}
	if value, ok := CoerceString(Type{Kind: Int}, "9223372036854775808"); ok {
		t.Fatalf("overflowing integer was accepted as %v", value)
	}
}

func TestUnionAndNullableCompatibility(t *testing.T) {
	nullable := Parse("int?")
	if nullable.Kind != Union || nullable.String() != "int|null" {
		t.Fatalf("nullable type = %#v", nullable)
	}
	if !Assignable(nullable, Type{Kind: Int}) || !Assignable(nullable, Type{Kind: Null}) {
		t.Fatal("nullable int must accept int and null")
	}
	if Assignable(nullable, Type{Kind: String}) {
		t.Fatal("nullable int must reject string")
	}
	if !Assignable(Parse("int|string"), Parse("int|string")) || Assignable(Parse("int"), Parse("int|string")) {
		t.Fatal("union source compatibility is not sound")
	}
}
