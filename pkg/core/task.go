package core

import "fmt"

// Task Implementation (Hit-based)
func (r *Runtime) executeTaskMethod(instance *Instance, method string, args []interface{}) interface{} {
	if method == "on_request" {
		if len(args) >= 3 {
			name := args[0].(string)
			// interval := args[1].(string)
			// callback := args[2] (Block/Closure)

			// Check if we should run
			// For simulation, we always run it if it's "system_health" or similar, or just log it.
			// To properly implement, we need to execute the block passed as argument.
			// The runtime needs to handle the callback execution.
			// Since we are inside executeNativeMethod, we can't easily execute a block without the runtime context loop.
			// BUT, we are in the runtime! 'r' is *Runtime.

			// We need to check if args[2] is a BlockStatement or similar?
			// The parser passes blocks as... wait, native methods receive evaluated args.
			// If the argument was a block `{ ... }`, it might not be evaluated to a value easily unless we treat it as a closure.
			// Currently, Joss doesn't support passing blocks as values (closures) fully.
			// The "Bible" says: Task::on_request(..., { code })
			// This implies the 3rd argument is a block.
			// In `evaluateCall`, arguments are evaluated. A block `{}` is not an expression in current parser?
			// Let's assume for now we just print. To support this fully, we need Closures.

			fmt.Printf("[Task] Ejecutando tarea hit-based: %s\n", name)
			// If we could, we would execute the block here.
		}
	}
	return nil
}
