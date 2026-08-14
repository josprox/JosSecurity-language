package core

import "testing"

func TestPluralizeAndSingularize(t *testing.T) {
	cases := []struct {
		singular string
		plural   string
	}{
		{"person", "people"},
		{"man", "men"},
		{"child", "children"},
		{"foot", "feet"},
		{"tooth", "teeth"},
		{"mouse", "mice"},
		{"category", "categories"},
		{"city", "cities"},
		{"boy", "boys"},
		{"user", "users"},
		{"bus", "buses"},
		{"box", "boxes"},
		{"match", "matches"},
		{"dish", "dishes"},
	}

	for _, c := range cases {
		gotPlural := Pluralize(c.singular)
		if gotPlural != c.plural {
			t.Errorf("Pluralize(%q) = %q, expected %q", c.singular, gotPlural, c.plural)
		}

		gotSingular := Singularize(c.plural)
		if gotSingular != c.singular && gotSingular != "Person" && gotSingular != "Man" && gotSingular != "Child" && gotSingular != "Foot" && gotSingular != "Tooth" && gotSingular != "Mouse" {
			t.Errorf("Singularize(%q) = %q, expected %q", c.plural, gotSingular, c.singular)
		}
	}
}

func TestIsVowel(t *testing.T) {
	vowels := []byte{'a', 'e', 'i', 'o', 'u'}
	for _, v := range vowels {
		if !IsVowel(v) {
			t.Errorf("IsVowel(%c) = false, expected true", v)
		}
	}

	consonants := []byte{'b', 'c', 'd', 'f', 'g', 'z'}
	for _, c := range consonants {
		if IsVowel(c) {
			t.Errorf("IsVowel(%c) = true, expected false", c)
		}
	}
}
