package parser

import "testing"

func TestLexerTracksLineAndColumn(t *testing.T) {
	lexer := NewLexer("$first = 1\n  $second = 2")
	for {
		token := lexer.NextToken()
		if token.Type == IDENT && token.Literal == "second" {
			if token.Line != 2 || token.Column != 4 {
				t.Fatalf("second position = %d:%d, want 2:4", token.Line, token.Column)
			}
			return
		}
		if token.Type == EOF {
			t.Fatal("second identifier not found")
		}
	}
}
