// Package value contains runtime value semantics that are independent from
// the AST evaluator and framework integrations.
package value

import "github.com/rivo/uniseg"

// StringIndex returns one user-perceived Unicode character (extended grapheme
// cluster) at index. It never returns a partial UTF-8 sequence, combining mark,
// variation selector or ZWJ fragment by itself.
func StringIndex(text string, index int64) (string, bool) {
	if index < 0 {
		return "", false
	}
	graphemes := uniseg.NewGraphemes(text)
	for current := int64(0); graphemes.Next(); current++ {
		if current == index {
			return graphemes.Str(), true
		}
	}
	return "", false
}
