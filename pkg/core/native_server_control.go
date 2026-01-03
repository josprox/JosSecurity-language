package core

import (
	"fmt"
	"net/http"
)

// ServerStartCallback allows avoiding import cycle with pkg/server
var ServerStartCallback func(fs http.FileSystem)

// Server Control Native Class
func (r *Runtime) executeServerControlMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "start":
		fmt.Println("[Server::start] Iniciando servidor web...")

		fs := GlobalFileSystem

		if ServerStartCallback != nil {
			ServerStartCallback(fs)
		} else {
			fmt.Println("[Server::start] Error: Server callback no registrado.")
		}

		return true
	}
	return nil
}
