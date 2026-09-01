package formatter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jossecurity/joss/pkg/parser"
)

type Formatter struct {
	tokens       []Token
	pos          int
	indent       int
	out          strings.Builder
	tokensOnLine int
	consecLines  int
	prevTok      Token
}

func FormatSource(src string) (string, error) {
	p := parser.NewParser(parser.NewLexer(src))
	_ = p.ParseProgram()
	if len(p.Errors()) > 0 {
		return "", fmt.Errorf("syntax error: %s", p.Errors()[0])
	}

	scanner := NewScanner(src)
	tokens := scanner.ScanAll()

	f := &Formatter{
		tokens: tokens,
		pos:    0,
		indent: 0,
	}

	return f.formatTokens(), nil
}

func (f *Formatter) peek() Token {
	if f.pos >= len(f.tokens) {
		return Token{Kind: TokEOF}
	}
	return f.tokens[f.pos]
}

func (f *Formatter) peekAt(offset int) Token {
	idx := f.pos + offset
	if idx >= len(f.tokens) || idx < 0 {
		return Token{Kind: TokEOF}
	}
	return f.tokens[idx]
}

func (f *Formatter) advance() Token {
	if f.pos >= len(f.tokens) {
		return Token{Kind: TokEOF}
	}
	tok := f.tokens[f.pos]
	f.pos++
	return tok
}

func (f *Formatter) writeIndent() {
	if f.tokensOnLine == 0 && f.indent > 0 {
		f.out.WriteString(strings.Repeat("    ", f.indent))
	}
}

func (f *Formatter) writeNewline() {
	f.out.WriteRune('\n')
	f.tokensOnLine = 0
	f.consecLines++
}

func (f *Formatter) formatTokens() string {
	var inGenericBrackets int // tracking <T> in types
	var blockDepth int        // tracking nesting inside { ... }
	var sourceConsecNewlines int

	for f.pos < len(f.tokens) {
		tok := f.advance()

		if tok.Kind == TokEOF {
			break
		}

		// Handle Newlines from original source
		if tok.Kind == TokNewline {
			sourceConsecNewlines++
			if sourceConsecNewlines >= 2 && f.consecLines < 2 && f.out.Len() > 0 {
				f.writeNewline()
			}
			continue
		}
		sourceConsecNewlines = 0

		// Handle Comments
		if tok.Kind == TokCommentLine || tok.Kind == TokCommentBlock {
			f.writeIndent()
			if f.tokensOnLine > 0 {
				f.out.WriteRune(' ')
			}
			f.out.WriteString(tok.Text)
			f.tokensOnLine++
			if tok.Kind == TokCommentLine {
				f.writeNewline()
			}
			f.prevTok = tok
			continue
		}

		// Block closing: }
		if tok.Kind == TokDelimiter && tok.Text == "}" {
			if f.indent > 0 {
				f.indent--
			}
			if blockDepth > 0 {
				blockDepth--
			}
			if f.tokensOnLine > 0 {
				f.writeNewline()
			}
			f.writeIndent()
			f.out.WriteString("}")
			f.tokensOnLine++
			f.consecLines = 0

			// If followed by : { (ternary block false branch), stay on same line
			next := f.peek()
			if next.Kind == TokDelimiter && next.Text == ":" && f.peekAt(1).Kind == TokDelimiter && f.peekAt(1).Text == "{" {
				f.out.WriteString(" : {")
				f.advance() // :
				f.advance() // {
				f.indent++
				blockDepth++
				f.writeNewline()
				f.prevTok = Token{Kind: TokDelimiter, Text: "{"}
				continue
			}

			// If followed by catch, stay on same line or space
			if next.Kind == TokIdent && next.Text == "catch" {
				f.out.WriteRune(' ')
			} else if next.Kind != TokDelimiter || (next.Text != ";" && next.Text != "," && next.Text != ")") {
				f.writeNewline()
			}
			f.prevTok = tok
			continue
		}

		// Block or Map opening: {
		if tok.Kind == TokDelimiter && tok.Text == "{" {
			isBlock := f.isBlockOpen()
			f.writeIndent()
			if f.tokensOnLine > 0 {
				f.out.WriteRune(' ')
			}
			f.out.WriteString("{")
			f.tokensOnLine++
			f.prevTok = tok
			if isBlock {
				f.indent++
				blockDepth++
				f.consecLines = 0
				f.writeNewline()
			} else {
				// Map or object literal inline
				f.consecLines = 0
			}
			continue
		}

		// Statement terminator: ;
		if tok.Kind == TokDelimiter && tok.Text == ";" {
			f.out.WriteString(";")
			f.tokensOnLine++
			f.consecLines = 0
			f.writeNewline()
			f.prevTok = tok
			continue
		}

		// Comma: ,
		if tok.Kind == TokDelimiter && tok.Text == "," {
			f.out.WriteString(",")
			f.tokensOnLine++
			f.prevTok = tok
			if f.peek().Kind == TokNewline {
				f.advance() // consume TokNewline
				f.writeNewline()
			}
			continue
		}

		// Ensure indentation at start of line
		f.writeIndent()

		// Generic bracket tracking: array<int>, map<string, User>
		if tok.Kind == TokOperator && tok.Text == "<" {
			if f.prevTok.Kind == TokIdent && (f.prevTok.Text == "array" || f.prevTok.Text == "map") {
				inGenericBrackets++
				f.out.WriteString("<")
				f.tokensOnLine++
				f.prevTok = tok
				f.consecLines = 0
				continue
			}
		}
		if tok.Kind == TokOperator && tok.Text == ">" && inGenericBrackets > 0 {
			inGenericBrackets--
			f.out.WriteString(">")
			f.tokensOnLine++
			f.prevTok = tok
			f.consecLines = 0
			continue
		}

		// Spacing before token (only if not first token on line)
		if f.tokensOnLine > 0 {
			if shouldSpaceBefore(tok, f.prevTok, inGenericBrackets) {
				f.out.WriteRune(' ')
			}
		}

		f.out.WriteString(tok.Text)
		f.tokensOnLine++
		f.consecLines = 0
		f.prevTok = tok
	}

	result := f.out.String()
	trimmed := strings.TrimRight(result, " \t\n\r")
	if len(trimmed) > 0 {
		return trimmed + "\n"
	}
	return ""
}

func (f *Formatter) isBlockOpen() bool {
	switch f.prevTok.Text {
	case ")", ":", "?", "do", "try", "else":
		return true
	}
	if f.prevTok.Kind == TokIdent {
		// Class Name { or func name() {
		return true
	}
	return false
}

func shouldSpaceBefore(curr, prev Token, inGenerics int) bool {
	if prev.Kind == TokDelimiter {
		if prev.Text == "(" || prev.Text == "[" || prev.Text == "{" {
			return false
		}
		if prev.Text == "," || prev.Text == ";" || prev.Text == ":" || prev.Text == "?" {
			return true
		}
	}

	if curr.Kind == TokDelimiter {
		if curr.Text == "(" {
			if prev.Kind == TokIdent {
				switch prev.Text {
				case "while", "foreach", "match", "catch", "if":
					return true
				default:
					return false
				}
			}
			if prev.Kind == TokVar {
				return false
			}
		}
		if curr.Text == "," || curr.Text == ";" || curr.Text == ")" || curr.Text == "]" {
			return false
		}
		if curr.Text == ":" {
			return false
		}
		if curr.Text == "{" {
			return true
		}
		if curr.Text == "?" {
			return true
		}
	}

	// Operators
	if curr.Kind == TokOperator {
		if curr.Text == "->" || curr.Text == "::" || curr.Text == "++" || curr.Text == "--" {
			return false
		}
		if inGenerics > 0 && (curr.Text == "<" || curr.Text == ">") {
			return false
		}
		if curr.Text == "!" {
			return false
		}
		return true
	}

	if prev.Kind == TokOperator {
		if prev.Text == "->" || prev.Text == "::" || prev.Text == "!" {
			return false
		}
		if inGenerics > 0 && (prev.Text == "<" || prev.Text == ">") {
			return false
		}
		return true
	}

	return true
}

func FormatFile(filePath string, write bool) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	formatted, err := FormatSource(string(data))
	if err != nil {
		return false, err
	}

	changed := formatted != string(data)
	if changed && write {
		if err := os.WriteFile(filePath, []byte(formatted), 0644); err != nil {
			return false, err
		}
	}
	return changed, nil
}

func FormatDirectory(root string, write bool, check bool) ([]string, error) {
	var unformatted []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" || name == ".cache" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".joss") {
			changed, formatErr := FormatFile(path, write)
			if formatErr != nil {
				return fmt.Errorf("error formatting %s: %w", path, formatErr)
			}
			if changed {
				unformatted = append(unformatted, path)
			}
		}
		return nil
	})

	return unformatted, err
}
