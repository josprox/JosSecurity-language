package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	semanticanalyzer "github.com/jossecurity/joss/pkg/analyzer"
	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/template"
)

func TestMigrationTableNameNormalizesDocumentedForms(t *testing.T) {
	for input, want := range map[string]string{
		"create_products":       "products",
		"create_products_table": "products",
		"product":               "products",
	} {
		got, err := migrationTableName(input)
		if err != nil {
			t.Fatalf("migrationTableName(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("migrationTableName(%q) = %q, want %q", input, got, want)
		}
	}
	if _, err := migrationTableName("../../products"); err == nil {
		t.Fatal("unsafe migration name was accepted")
	}
}

func TestMigrationAndCRUDGeneratorsWorkTogether(t *testing.T) {
	root := t.TempDir()
	template.CreateBibleProject(root)
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(previous) }()

	if err := createMigration("create_products_table"); err != nil {
		t.Fatal(err)
	}
	migrations, err := filepath.Glob(filepath.Join("app", "database", "migrations", "*.joss"))
	if err != nil || len(migrations) != 1 {
		t.Fatalf("generated migrations = %v, err = %v", migrations, err)
	}
	migration, err := os.ReadFile(migrations[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(migration), "class CreateProductsTable") ||
		!strings.Contains(string(migration), `Schema::create("products"`) {
		t.Fatalf("migration uses an incorrect table or class:\n%s", migration)
	}

	db, err := sql.Open("sqlite", filepath.Join(root, "database.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE js_categories (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE js_products (id INTEGER PRIMARY KEY, name TEXT NOT NULL, category_id INTEGER, legacy_id INTEGER, created_at TIMESTAMP, updated_at TIMESTAMP)`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	_ = db.Close()

	if err := createCRUD("products"); err != nil {
		t.Fatal(err)
	}
	controller, err := os.ReadFile(filepath.Join("app", "controllers", "ProductController.joss"))
	if err != nil {
		t.Fatal(err)
	}
	controllerText := string(controller)
	if strings.Contains(controllerText, "Request::except") ||
		!strings.Contains(controllerText, `"name": Request::input("name")`) ||
		!strings.Contains(controllerText, `"category_id": Request::input("category_id")`) ||
		!strings.Contains(controllerText, `"legacy_id": Request::input("legacy_id")`) {
		t.Fatalf("controller does not whitelist writable fields:\n%s", controllerText)
	}
	if !strings.Contains(controllerText, `->leftJoin("categories"`) || strings.Contains(controllerText, `->leftJoin("legacies"`) {
		t.Fatalf("controller did not distinguish a real relation from an orphan _id column:\n%s", controllerText)
	}
	if _, err := os.Stat(filepath.Join("app", "models", "Category.joss")); err != nil {
		t.Fatalf("related model was not generated: %v", err)
	}
	routes, err := os.ReadFile("routes.joss")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routes), `Router::post("/product/delete/{id}"`) ||
		strings.Contains(string(routes), `Router::get("/product/delete/{id}"`) {
		t.Fatalf("generated deletion route is not POST-only:\n%s", routes)
	}

	if err := createCRUD("products"); err != nil {
		t.Fatal(err)
	}
	routes, _ = os.ReadFile("routes.joss")
	if count := strings.Count(string(routes), "// CRUD Routes for Product"); count != 1 {
		t.Fatalf("CRUD route injection is not idempotent: %d blocks", count)
	}

	units, parseDiagnostics := semanticanalyzer.LoadProject("main.joss", "app")
	if len(parseDiagnostics) != 0 {
		t.Fatalf("generated project parse diagnostics: %#v", parseDiagnostics)
	}
	if report := core.AnalyzeSourceUnits(units); report.HasErrors() {
		t.Fatalf("generated project semantic diagnostics: %#v", report.Diagnostics)
	}
}

func TestCRUDRejectsUnsafeTableName(t *testing.T) {
	if err := createCRUD(`products; DROP TABLE users`); err == nil {
		t.Fatal("unsafe table name was accepted")
	}
}
