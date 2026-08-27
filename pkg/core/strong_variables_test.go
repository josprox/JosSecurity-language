package core

import "testing"

func TestRuntimeRequiresExplicitMixedForDynamicParameter(t *testing.T) {
	mustPanicWithType(t, `public func identity($value) { return $value }
$result = identity(1)`, "ImplicitMixedParameter")
	mustNotPanic(t, `public func identity(mixed $value) { return $value }
$result = identity(1)`)
}
