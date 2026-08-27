package analyzer

import "testing"

func TestAnalyzerEnforcesPrivateAndProtectedMembers(t *testing.T) {
	items := analyzeSource(t, `public class Vault {
    private int $secret = 7
    private func reveal(): int { return $this->secret }
}
public func leak(Vault $vault): int { return $vault->reveal() }`, NewEnvironment())
	if !hasCode(items, "JOSS-ACCESS-002") {
		t.Fatalf("expected access diagnostic, got %#v", items)
	}
}

func TestAnalyzerAllowsProtectedAccessFromSubclass(t *testing.T) {
	items := analyzeSource(t, `public class Base {
    protected int $value = 7
    protected func read(): int { return $this->value }
}
public class Child extends Base {
    public func expose(): int { return $this->read() }
}`, NewEnvironment())
	if hasCode(items, "JOSS-ACCESS-002") {
		t.Fatalf("unexpected access diagnostic: %#v", items)
	}
}
