package frame

import "testing"

func TestPrimitiveValuesRoundTripWithoutDynamicStorage(t *testing.T) {
	tests := []struct {
		value interface{}
		kind  ValueKind
	}{
		{nil, Null}, {int64(42), Int}, {3.5, Float}, {true, Bool}, {"joss", String},
	}
	for _, test := range tests {
		value := FromInterface(test.value)
		if value.Kind != test.kind || value.Dynamic != nil {
			t.Fatalf("FromInterface(%#v) = %#v", test.value, value)
		}
		if got := value.Interface(); got != test.value {
			t.Fatalf("round trip %#v = %#v", test.value, got)
		}
	}
}

func TestDynamicValueRetainsObjects(t *testing.T) {
	object := map[string]interface{}{"id": int64(1)}
	value := FromInterface(object)
	if value.Kind != Dynamic || value.Dynamic == nil {
		t.Fatalf("dynamic value = %#v", value)
	}
}
