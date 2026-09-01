package analyzer

import (
	"testing"
)

func TestTypedCollectionsAnalyzerInference(t *testing.T) {
	items := analyzeSource(t, `
array<int> $numbers = [1, 2, 3]
int $_first = $numbers[0]
`, NewEnvironment())
	if len(items) > 0 {
		t.Fatalf("expected no errors for valid typed array index access, got %#v", items)
	}

	badItems := analyzeSource(t, `
array<int> $_nums = ["hello"]
`, NewEnvironment())
	if !hasCode(badItems, "JOSS-TYPE-002") {
		t.Fatalf("expected type mismatch error assigning array<string> to array<int>, got %#v", badItems)
	}
}

func TestTypeNarrowingInTernaryBlocks(t *testing.T) {
	items := analyzeSource(t, `
public class User {
    public string $name = "Alice"
}
public func getUser(): User|null {
    return null
}
public func test(): string {
    User|null $u = getUser()
    return $u != null ? {
        return $u->name
    } : {
        return "default"
    }
}
`, NewEnvironment())
	if len(items) > 0 {
		t.Fatalf("expected clean analysis with type narrowing, got diagnostics: %#v", items)
	}
}

func TestTypeNarrowingEqualityBranch(t *testing.T) {
	items := analyzeSource(t, `
public class Profile {
    public string $email = "test@example.com"
}
public func getProfile(): Profile|null {
    return null
}
public func run(): string {
    Profile|null $p = getProfile()
    return $p == null ? {
        return "no email"
    } : {
        return $p->email
    }
}
`, NewEnvironment())
	if len(items) > 0 {
		t.Fatalf("expected clean analysis with equality narrowing, got diagnostics: %#v", items)
	}
}
