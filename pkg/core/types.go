package core

import (
	"database/sql"

	"github.com/jossecurity/joss/pkg/parser"
)

// Runtime manages the execution environment of a Joss program
type Runtime struct {
	Env               map[string]string
	Variables         map[string]interface{}
	VarTypes          map[string]string // For strict typing
	Classes           map[string]*parser.ClassStatement
	Functions         map[string]*parser.MethodStatement
	DB                *sql.DB
	Routes            map[string]map[string]interface{} // HTTP Method -> Path -> Handler
	CurrentMiddleware []string
}

// Instance represents an instance of a class
type Instance struct {
	Class  *parser.ClassStatement
	Fields map[string]interface{}
}

// BoundMethod represents a method bound to an instance
type BoundMethod struct {
	Method   *parser.MethodStatement
	Instance *Instance
}

// Future represents an asynchronous computation
type Future struct {
	done   chan bool
	result interface{}
	err    error
}

// Wait blocks until the Future completes and returns the result
func (f *Future) Wait() interface{} {
	<-f.done
	if f.err != nil {
		panic(f.err)
	}
	return f.result
}

// Cout represents standard output stream
type Cout struct{}

func (c *Cout) String() string { return "cout" }

// Cin represents standard input stream
type Cin struct{}

func (c *Cin) String() string { return "cin" }
