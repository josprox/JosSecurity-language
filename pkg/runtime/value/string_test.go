package value

import "testing"

func TestStringIndexUsesGraphemeClusters(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		index int64
		want  string
	}{
		{"ascii", "Joss", 1, "o"},
		{"accent", "México", 1, "é"},
		{"combining accent", "e\u0301cole", 0, "e\u0301"},
		{"cjk", "语言", 1, "言"},
		{"emoji", "🙂!", 0, "🙂"},
		{"compound emoji", "👩‍💻!", 0, "👩‍💻"},
		{"flag", "🇲🇽!", 0, "🇲🇽"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := StringIndex(test.text, test.index)
			if !ok || got != test.want {
				t.Fatalf("StringIndex(%q, %d) = %q, %v; want %q", test.text, test.index, got, ok, test.want)
			}
		})
	}
}

func TestStringIndexRejectsInvalidBounds(t *testing.T) {
	for _, index := range []int64{-1, 1} {
		if value, ok := StringIndex("é", index); ok {
			t.Fatalf("StringIndex at %d returned %q", index, value)
		}
	}
}
