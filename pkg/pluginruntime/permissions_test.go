package pluginruntime

import (
	"strings"
	"testing"

	"github.com/jossecurity/joss/pkg/plugincompiler/ir"
)

type dummyHost struct{}

func (d *dummyHost) CallHostFunction(name string, args []interface{}) (interface{}, error) {
	return "host_ok", nil
}

func TestJPBCVMPermissionEnforcement(t *testing.T) {
	// Create module calling host function "http_get"
	module := &JPBCModule{
		MajorVersion: 1,
		MinorVersion: 0,
		ConstantPool: []interface{}{"http_get"},
		Functions: map[string]*JPBCFunction{
			"fetchData": {
				Name: "fetchData",
				Instructions: []JPBCInstruction{
					{Op: ir.OpCallStatic, ConstIdx: 0}, // call "http_get"
					{Op: ir.OpReturn},
				},
			},
		},
	}

	// Case 1: Plugin WITHOUT required "network.http" permission -> SHOULD FAIL
	guardDenied := NewPermissionGuard([]string{"filesystem.read"})
	vmDenied := NewJPBCVM(module, guardDenied, &dummyHost{})
	_, err := vmDenied.Execute("fetchData", nil)
	if err == nil {
		t.Errorf("Expected permission denied error for fetchData without network.http permission, got nil error")
	} else if !strings.Contains(err.Error(), "permiso denegado") {
		t.Errorf("Expected 'permiso denegado' error message, got: %v", err)
	}

	// Case 2: Plugin WITH required "network.http" permission -> SHOULD SUCCEED
	guardAllowed := NewPermissionGuard([]string{"network.http"})
	vmAllowed := NewJPBCVM(module, guardAllowed, &dummyHost{})
	res, err := vmAllowed.Execute("fetchData", nil)
	if err != nil {
		t.Errorf("Unexpected error for authorized fetchData call: %v", err)
	}
	if res != "host_ok" {
		t.Errorf("Expected 'host_ok', got %v", res)
	}
}
