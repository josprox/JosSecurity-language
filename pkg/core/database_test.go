package core

import (
	"testing"
)

func TestGranDBCallableWhere(t *testing.T) {
	r := NewRuntime()
	inst := &Instance{
		Class:  nil,
		Fields: make(map[string]interface{}),
	}

	// Test case-insensitive orWhere and orWhereLike
	r.executeGranDBMethod(inst, "table", []interface{}{"pub_packages"})
	r.executeGranDBMethod(inst, "where", []interface{}{"is_deprecated", 0})
	r.executeGranDBMethod(inst, "orWhereLike", []interface{}{"name", "joss"})

	wheres := inst.Fields["_wheres"].([]string)
	if len(wheres) != 2 {
		t.Fatalf("Se esperaban 2 condiciones where, se obtuvieron: %d", len(wheres))
	}

	sqlStr, _ := r.buildSelectQuery(inst, "*")
	expected := "SELECT * FROM `js_pub_packages` WHERE `is_deprecated` = ? OR `name` LIKE ?"
	if sqlStr != expected {
		t.Errorf("SQL generado incorrecto.\nEsperado: %s\nObtenido: %s", expected, sqlStr)
	}
}

func TestGranDBCountPreservesWhere(t *testing.T) {
	r := NewRuntime()
	inst := &Instance{
		Class:  nil,
		Fields: make(map[string]interface{}),
	}

	r.executeGranDBMethod(inst, "table", []interface{}{"pub_packages"})
	r.executeGranDBMethod(inst, "where", []interface{}{"is_deprecated", 0})
	r.executeGranDBMethod(inst, "whereLike", []interface{}{"name", "backup"})

	// Verify buildSelectQuery before count
	sqlBefore, _ := r.buildSelectQuery(inst, "COUNT(*)")
	expectedBefore := "SELECT COUNT(*) FROM `js_pub_packages` WHERE `is_deprecated` = ? AND `name` LIKE ?"
	if sqlBefore != expectedBefore {
		t.Errorf("SQL de count incorrecto.\nEsperado: %s\nObtenido: %s", expectedBefore, sqlBefore)
	}

	// Verify wheres are preserved for subsequent get()
	wheres := inst.Fields["_wheres"].([]string)
	if len(wheres) != 2 {
		t.Fatalf("Wheres se borraron tras count! Se esperaban 2, se obtuvieron %d", len(wheres))
	}
}
