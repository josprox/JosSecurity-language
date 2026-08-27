package core

import "testing"

func TestRuntimeRejectsExternalPrivateMemberAccess(t *testing.T) {
	mustPanicWithType(t, `public class Vault {
    private int $secret = 7
    private func reveal(): int { return $this->secret }
}
$vault = new Vault()
$value = $vault->reveal()`, "AccessViolation")
}

func TestRuntimeAllowsProtectedMemberFromSubclass(t *testing.T) {
	mustNotPanic(t, `public class Base {
    protected int $value = 7
    protected func read(): int { return $this->value }
}
public class Child extends Base {
    public func expose(): int { return $this->read() }
}
$child = new Child()
$value = $child->expose()`)
}
