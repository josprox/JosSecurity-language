package parser

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	INT    = "INT"    // 1343456
	STRING = "STRING" // "foobar"

	// Operators and delimiters
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT          = "<"
	GT          = ">"
	EQ          = "=="
	NOT_EQ      = "!="
	LTE         = "<="
	GTE         = ">="
	SHIFT_LEFT  = "<<"
	SHIFT_RIGHT = ">>"

	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	QUESTION  = "?"
	NEWLINE   = "NEWLINE"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"
	DOT      = "."

	// Keywords
	FUNCTION  = "FUNCTION"
	VAR       = "VAR" // $
	TRUE      = "TRUE"
	FALSE     = "FALSE"
	IF        = "IF"
	ELSE      = "ELSE"
	RETURN    = "RETURN"
	PRINT     = "PRINT"
	ECHO      = "ECHO"
	CLASS     = "CLASS"
	INIT      = "INIT"
	NAMESPACE = "NAMESPACE"
	IMPORT    = "IMPORT"
	NEW       = "NEW"
	FOREACH   = "FOREACH"
	AS        = "AS"
	THIS      = "THIS"
	ISSET     = "ISSET"
	EMPTY     = "EMPTY"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"true":      TRUE,
	"false":     FALSE,
	"if":        IF,
	"else":      ELSE,
	"return":    RETURN,
	"class":     CLASS,
	"Init":      INIT,
	"Namespace": NAMESPACE,
	"Import":    IMPORT,
	"new":       NEW,
	"foreach":   FOREACH,
	"as":        AS,
	"function":  FUNCTION,
	"this":      THIS,
	"echo":      ECHO,
	"print":     PRINT,
	"isset":     ISSET,
	"empty":     EMPTY,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
