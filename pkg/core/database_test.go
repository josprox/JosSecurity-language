package core

import (
	"database/sql"
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

func TestGranDBAcceptsCanonicalMapInsertAndRejectsParallelArrays(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE items (name TEXT, amount INTEGER)`); err != nil {
		t.Fatal(err)
	}

	r := NewRuntime()
	r.DB = db
	r.Env = map[string]string{"DB": "sqlite", "PREFIX": ""}
	instance := &Instance{Fields: make(map[string]interface{})}
	r.executeGranDBMethod(instance, "table", []interface{}{"items"})

	removed := r.executeInsertMethod(instance, []interface{}{[]interface{}{"name", "amount"}, []interface{}{"old", int64(1)}}, false)
	if removed != false {
		t.Fatalf("parallel-array insert returned %v, want false", removed)
	}
	if inserted := r.executeInsertMethod(instance, []interface{}{map[string]interface{}{"name": "current", "amount": int64(2)}}, false); inserted != true {
		t.Fatalf("map insert returned %v, want true", inserted)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("row count = %d, err = %v", count, err)
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

func TestGranDBTableResetsQueryStateForNewQuery(t *testing.T) {
	r := NewRuntime()
	inst := &Instance{
		Class:  nil,
		Fields: make(map[string]interface{}),
	}

	// 1st query: sync_change_log with user_id and client_change_id
	r.executeGranDBMethod(inst, "table", []interface{}{"sync_change_log"})
	r.executeGranDBMethod(inst, "where", []interface{}{"user_id", 60})
	r.executeGranDBMethod(inst, "where", []interface{}{"client_change_id", "abc-123"})

	wheres := inst.Fields["_wheres"].([]string)
	if len(wheres) != 2 {
		t.Fatalf("Expected 2 where clauses, got %d", len(wheres))
	}

	// 2nd query: switch table to user_recent_plays
	r.executeGranDBMethod(inst, "table", []interface{}{"user_recent_plays"})
	r.executeGranDBMethod(inst, "where", []interface{}{"user_id", 60})
	r.executeGranDBMethod(inst, "where", []interface{}{"track_id", 79})

	sqlStr, bindings := r.buildSelectQuery(inst, "*")
	expected := "SELECT * FROM `js_user_recent_plays` WHERE `user_id` = ? AND `track_id` = ?"
	if sqlStr != expected {
		t.Errorf("SQL incorrecto tras cambiar de tabla.\nEsperado: %s\nObtenido: %s", expected, sqlStr)
	}
	if len(bindings) != 2 || bindings[0] != 60 || bindings[1] != 79 {
		t.Errorf("Bindings incorrectos tras cambiar de tabla: %v", bindings)
	}
}
