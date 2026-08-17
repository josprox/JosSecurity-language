package core

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
)

func (r *Runtime) callBuiltinAsync(name string, args []interface{}) (interface{}, bool) {
	switch name {
	case "async":
		if len(args) == 1 {
			future := &Future{
				done: make(chan bool),
			}
			argVal := args[0]
			newR := r.Fork() // Fork BEFORE starting the goroutine to avoid race
			go func() {
				defer func() {
					if p := recover(); p != nil {
						if rp, ok := p.(*ReturnPanic); ok {
							future.result = rp.Value
						} else {
							fmt.Printf("[ASYNC PANIC] %v\n", p)
							future.err = fmt.Errorf("%v", p)
						}
					}
					close(future.done)
				}()

				if fn, ok := argVal.(*parser.FunctionLiteral); ok {
					future.result = newR.executeBlock(fn.Body)
				} else if blk, ok := argVal.(*parser.BlockStatement); ok {
					future.result = newR.executeBlock(blk)
				} else {
					future.result = argVal
				}
			}()
			return future, true
		}
		return nil, true

	case "await":
		if len(args) == 1 {
			if future, ok := args[0].(*Future); ok {
				return future.Wait(), true
			}
		}
		return nil, true

	case "make_chan":
		size := 0
		if len(args) > 0 {
			if s, ok := args[0].(int64); ok {
				size = int(s)
			}
		}
		return &Channel{Ch: make(chan interface{}, size)}, true

	case "close":
		if len(args) == 1 {
			if ch, ok := args[0].(*Channel); ok {
				close(ch.Ch)
				return nil, true
			}
		}
		return nil, true

	case "send":
		if len(args) == 2 {
			if ch, ok := args[0].(*Channel); ok {
				ch.Ch <- args[1]
				return nil, true
			}
		}
		return nil, true

	case "recv":
		if len(args) == 1 {
			if ch, ok := args[0].(*Channel); ok {
				val, ok := <-ch.Ch
				if !ok {
					return nil, true
				}
				return val, true
			}
		}
		return nil, true
	}

	return nil, false
}
