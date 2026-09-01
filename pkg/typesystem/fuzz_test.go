package typesystem

import "testing"

func FuzzTypeParser(f *testing.F) {
	seeds := []string{
		"int",
		"string",
		"bool",
		"float",
		"array",
		"map",
		"array<int>",
		"map<string, User>",
		"User|null",
		"array<int|string>",
		"int?",
		"mixed",
		"unknown",
		"<<<>>>|||???",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Must not panic on arbitrary malformed type strings
		parsed := Parse(input)
		_ = parsed.String()
		_ = Assignable(parsed, parsed)
	})
}
