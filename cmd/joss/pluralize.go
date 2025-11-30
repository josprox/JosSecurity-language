package main

import "strings"

func pluralize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	lower := strings.ToLower(s)

	// Irregular words
	irregular := map[string]string{
		"person": "people",
		"man":    "men",
		"child":  "children",
		"foot":   "feet",
		"tooth":  "teeth",
		"mouse":  "mice",
	}

	if val, ok := irregular[lower]; ok {
		return val
	}

	// Ends in 'y' preceded by consonant -> 'ies'
	if strings.HasSuffix(lower, "y") && len(lower) > 1 {
		lastChar := lower[len(lower)-1]
		secondLast := lower[len(lower)-2]
		if lastChar == 'y' && !isVowel(secondLast) {
			return s[:len(s)-1] + "ies"
		}
	}

	// Ends in s, x, z, ch, sh -> 'es'
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") || strings.HasSuffix(lower, "z") || strings.HasSuffix(lower, "ch") || strings.HasSuffix(lower, "sh") {
		return s + "es"
	}

	return s + "s"
}

func singularize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	lower := strings.ToLower(s)

	// Irregular words (reverse)
	irregular := map[string]string{
		"people":   "person",
		"men":      "man",
		"children": "child",
		"feet":     "foot",
		"teeth":    "tooth",
		"mice":     "mouse",
	}

	if val, ok := irregular[lower]; ok {
		// Try to preserve case? For now return lowercase or title case based on input?
		// Simple return is fine for model names usually
		return strings.Title(val)
	}

	// 'ies' -> 'y'
	if strings.HasSuffix(lower, "ies") {
		return s[:len(s)-3] + "y"
	}

	// 'es' -> '' (for s, x, z, ch, sh)
	// This is tricky because 'es' can be just 's' appended to 'e'.
	// Simple heuristic: if ends in 'es', try removing 'es'.
	// But 'boxes' -> 'box', 'buses' -> 'bus'.
	// 'lives' -> 'life' (hard).

	// For our specific case of 'categories' -> 'category', the 'ies' rule handles it.
	// For 'products' -> 'product', the 's' rule handles it.

	if strings.HasSuffix(lower, "s") {
		return s[:len(s)-1]
	}

	return s
}

func isVowel(c byte) bool {
	return c == 'a' || c == 'e' || c == 'i' || c == 'o' || c == 'u'
}
