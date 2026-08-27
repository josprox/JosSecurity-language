package core

import "testing"

func TestMutableReferenceUpdatesCallerAndSupportsRecursion(t *testing.T) {
	runtime := executeTypingSource(t, `public func reduce(ref int $value): int {
    ($value <= 0) ? { return $value } : {}
    $value = $value - 1
    return reduce(ref $value)
}
$count = 4
$result = reduce(ref $count)`)
	if got := runtime.Variables["count"]; got != int64(0) {
		t.Fatalf("caller count = %#v, want 0", got)
	}
	if got := runtime.Variables["result"]; got != int64(0) {
		t.Fatalf("result = %#v, want 0", got)
	}
}

func TestMutableReferenceRequiresExactTypeAtRuntime(t *testing.T) {
	mustPanicWithType(t, `public func overwrite(ref int $value): int { $value = 1 return $value }
float $amount = 1.5
overwrite(ref $amount)`, "ReferenceTypeError")
}

func TestMutableReferenceCannotEscapeToNativeFunction(t *testing.T) {
	mustPanicWithType(t, `$value = 1
print(ref $value)`, "ReferenceEscape")
}
