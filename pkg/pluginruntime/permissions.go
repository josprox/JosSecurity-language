package pluginruntime

import (
	"strings"
)

// PermissionGuard valida y gestiona permisos solicitados por plugins.
type PermissionGuard struct {
	Declared map[string]bool
	Granted  map[string]bool
}

// NewPermissionGuard crea un PermissionGuard a partir de los permisos declarados en metadata.
func NewPermissionGuard(declared []string) *PermissionGuard {
	guard := &PermissionGuard{
		Declared: make(map[string]bool, len(declared)),
		Granted:  make(map[string]bool, len(declared)),
	}
	for _, p := range declared {
		clean := strings.TrimSpace(p)
		if clean != "" {
			guard.Declared[clean] = true
			guard.Granted[clean] = true // Por defecto, se conceden los permisos declarados válidos
		}
	}
	return guard
}

// HasPermission verifica si el plugin tiene concedido un permiso o su wildcard (ej: network.*).
func (g *PermissionGuard) HasPermission(permission string) bool {
	if g == nil {
		return false
	}
	if g.Granted[permission] {
		return true
	}
	// Soporte para wildcards ej. "network.*" -> "network.http"
	if idx := strings.Index(permission, "."); idx != -1 {
		prefixWildcard := permission[:idx] + ".*"
		if g.Granted[prefixWildcard] {
			return true
		}
	}
	return false
}

// RevokePermission revoca un permiso especifico en tiempo de ejecucion.
func (g *PermissionGuard) RevokePermission(permission string) {
	if g != nil {
		delete(g.Granted, permission)
	}
}
