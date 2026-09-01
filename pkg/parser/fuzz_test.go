package parser

import "testing"

func FuzzParser(f *testing.F) {
	seeds := []string{
		`$x = 10`,
		`public func test(int $a): int { return $a + 1 }`,
		`public class User { public string $name = "Alice" }`,
		`array<int> $items = [1, 2, 3]`,
		`$val != null ? { return 1 } : { return 0 }`,
		`while ($i < 10) { $i++ }`,
		`try { throw new Exception("err") } catch ($e) { return 0 }`,
	}
	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		input := string(data)
		lexer := NewLexer(input)
		p := NewParser(lexer)
		// Must not panic on arbitrary malformed inputs
		_ = p.ParseProgram()
	})
}
