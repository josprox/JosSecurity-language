package formatter

import (
	"testing"
)

func TestFormatterIdempotence(t *testing.T) {
	sources := []string{
		`public func test(int $a): int {
    return $a + 1;
}`,
		`public class User {
    public string $name = "Alice";
    public func greet(): string {
        return "Hello " . $this->name;
    }
}`,
		`$isValid ? {
    save();
} : {
    reject();
}`,
		`array<int> $items = [1, 2, 3];
map $data = {"key": 5};`,
		`// Leading comment
public func compute(): int {
    // Inner comment
    int $x = 10;
    int $y = 20;
    return $x + $y;
}`,
	}

	for _, src := range sources {
		firstPass, err := FormatSource(src)
		if err != nil {
			t.Fatalf("first pass error: %v", err)
		}
		secondPass, err := FormatSource(firstPass)
		if err != nil {
			t.Fatalf("second pass error: %v", err)
		}
		if firstPass != secondPass {
			t.Fatalf("formatter is not idempotent!\nFirst pass:\n%s\nSecond pass:\n%s", firstPass, secondPass)
		}
	}
}

func TestFormatterNormalizesSpaces(t *testing.T) {
	input := `public   func   add( int $a,int $b ):int{return $a+$b;}`
	expected := `public func add(int $a, int $b): int {
    return $a + $b;
}
`
	got, err := FormatSource(input)
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	if got != expected {
		t.Fatalf("formatting mismatch:\nGot:\n%q\nWant:\n%q", got, expected)
	}
}

func TestFormatterPreservesComments(t *testing.T) {
	input := `// Global comment
public func run(): void {
    // Step 1
    int $step = 1;
}
`
	got, err := FormatSource(input)
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	if got != input {
		t.Fatalf("comments corrupted:\nGot:\n%s\nWant:\n%s", got, input)
	}
}

func TestFormatterMatchAndChaining(t *testing.T) {
	input := `public func handle(int $status): string {
    return match ($status) {
        200 => "ok",
        404 => "not found",
        default => "unknown",
    };
}
`
	got, err := FormatSource(input)
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	if got != input {
		t.Fatalf("match formatting error:\nGot:\n%s\nWant:\n%s", got, input)
	}
}

func TestFormatterBlockTernary(t *testing.T) {
	input := `$active ? {
    activate();
} : {
    deactivate();
}
`
	got, err := FormatSource(input)
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	if got != input {
		t.Fatalf("block ternary formatting error:\nGot:\n%s\nWant:\n%s", got, input)
	}
}
