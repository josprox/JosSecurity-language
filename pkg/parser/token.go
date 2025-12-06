package parser

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	INT    = "INT"    // 1343456
	FLOAT  = "FLOAT"  // 12.34
	STRING = "STRING" // "foobar"

	// Operators and delimiters
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	PERCENT  = "%"

	LT          = "<"
	GT          = ">"
	EQ          = "=="
	NOT_EQ      = "!="
	LTE         = "<="
	GTE         = ">="
	SHIFT_LEFT  = "<<"
	SHIFT_RIGHT = ">>"
	AND         = "&&"
	OR          = "||"
	INCREMENT   = "++"

	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	QUESTION  = "?"
	NEWLINE   = "NEWLINE"

	LPAREN        = "("
	RPAREN        = ")"
	LBRACE        = "{"
	RBRACE        = "}"
	LBRACKET      = "["
	RBRACKET      = "]"
	DOT           = "."
	ARROW         = "->"
	DOUBLE_COLON  = "::"
	PIPE          = "|>"
	NULL_COALESCE = "??"

	// Keywords
	FUNCTION = "FUNCTION"
	VAR      = "VAR" // $
	TRUE     = "TRUE"
	FALSE    = "FALSE"

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
	// Control Structures
	WHILE   = "WHILE"
	DO      = "DO"
	TRY     = "TRY"
	CATCH   = "CATCH"
	THROW   = "THROW"
	EXTENDS = "EXTENDS"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
}

var keywords = map[string]TokenType{
	"true":  TRUE,
	"false": FALSE,

	"return":    RETURN,
	"class":     CLASS,
	"Init":      INIT,
	"Namespace": NAMESPACE,
	"Import":    IMPORT,
	"new":       NEW,
	"foreach":   FOREACH,
	"as":        AS,
	"function":  FUNCTION,
	"func":      FUNCTION,
	"this":      THIS,
	"echo":      ECHO,
	"print":     PRINT,
	"isset":     ISSET,
	"empty":     EMPTY,
	"while":     WHILE,
	"do":        DO,
	"try":       TRY,
	"catch":     CATCH,
	"throw":     THROW,
	"extends":   EXTENDS,
	"@import":   IMPORT,
	"import":    IMPORT,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
