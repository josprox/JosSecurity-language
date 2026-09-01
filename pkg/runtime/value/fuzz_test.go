package value

import "testing"

func FuzzGraphemeIndex(f *testing.F) {
	seeds := []string{
		"hello",
		"México",
		"école",
		"👩‍💻",
		"🇲🇽",
		"🙂🎉🚀",
		"\x00\xff\xfe",
	}
	for _, seed := range seeds {
		f.Add(seed, int64(0))
		f.Add(seed, int64(1))
		f.Add(seed, int64(-1))
		f.Add(seed, int64(1000))
	}

	f.Fuzz(func(t *testing.T, s string, index int64) {
		// Must not panic with arbitrary unicode inputs and index values
		_, _ = StringIndex(s, index)
		_ = StringGraphemeCount(s)
	})
}
