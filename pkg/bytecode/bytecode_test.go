package bytecode

import (
	"testing"

	"github.com/jossecurity/joss/pkg/parser"
)

func TestBytecodeEncodeDecode(t *testing.T) {
	input := `
public class Calculator {
    public func add(int $a, int $b) {
        return $a + $b;
    }
}

int $x = 10;
string $msg = "hello";
int $y = 20, $z = 30;

($x < $y) ? {
    print($msg);
} : {}

foreach ([1, 2, 3] as $item) {
    print($item);
}
`
	l := parser.NewLexer(input)
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	encoded, err := Encode(prog)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !IsBytecode(encoded) {
		t.Fatalf("IsBytecode returned false on encoded payload")
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(decoded.Statements) != len(prog.Statements) {
		t.Errorf("Decoded statement count mismatch: got %d, want %d", len(decoded.Statements), len(prog.Statements))
	}
}

func TestLegacyUncompressedBytecodeIsRejected(t *testing.T) {
	legacy := append([]byte{'J', 'O', 'S', 'S', 'B', 'C', '2', 0}, []byte("legacy")...)
	if IsBytecode(legacy) {
		t.Fatal("legacy uncompressed bytecode must not be recognized")
	}
	if _, err := Decode(legacy); err == nil {
		t.Fatal("legacy uncompressed bytecode must be rejected")
	}
}
