package typesystem

import "testing"

func TestCanonicalAliasesAndCompatibility(t *testing.T) {
	if got := Parse("integer"); got.Kind != Int {
		t.Fatalf("integer alias = %s, want int", got.String())
	}
	if got := Parse("boolean"); got.Kind != Bool {
		t.Fatalf("boolean alias = %s, want bool", got.String())
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
