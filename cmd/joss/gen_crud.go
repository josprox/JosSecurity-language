package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

func createCRUD(tableName string) {
	fmt.Printf("Generating CRUD for table '%s'...\n", tableName)

	// 1. Connect to DB
	dbType, dbPath, dbHost, dbUser, dbPass, dbName, prefix := loadEnvConfig()

	var db *sql.DB
	var err error

	if dbType == "sqlite" {
		db, err = sql.Open("sqlite", dbPath)
	} else {
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPass, dbHost, dbName)
		db, err = sql.Open("mysql", dsn)
	}

	if err != nil {
		fmt.Printf("Error connecting to DB: %v\n", err)
		return
	}
	defer db.Close()

	// 2. Inspect Schema
	cols, err := getColumns(db, dbType, tableName)
	if err != nil {
		fmt.Printf("Error inspecting table: %v\n", err)
		return
	}

	// If table not found and doesn't start with prefix, try adding prefix
	if len(cols) == 0 && !strings.HasPrefix(tableName, prefix) {
		prefixedName := prefix + tableName
		fmt.Printf("Table '%s' not found. Trying '%s'...\n", tableName, prefixedName)
		cols, err = getColumns(db, dbType, prefixedName)
		if err != nil {
			fmt.Printf("Error inspecting table: %v\n", err)
			return
		}
		if len(cols) > 0 {
			tableName = prefixedName
			fmt.Printf("Found table '%s'. Using it.\n", tableName)
		}
	}

	if len(cols) == 0 {
		fmt.Printf("Table '%s' not found or empty.\n", tableName)
		return
	}
	// 3. Analyze Relations
	var relations []Relation
	for _, c := range cols {
		fmt.Printf("Inspecting column: '%s'\n", c.Name)
		if strings.HasSuffix(c.Name, "_id") {
			fmt.Printf("  -> Found relation for %s\n", c.Name)
			// Infer relation
			baseName := strings.TrimSuffix(c.Name, "_id")
			_, _, _, _, _, _, prefix := loadEnvConfig()
			relatedTable := prefix + strings.ToLower(pluralize(baseName)) // Convention: js_users

			// Smartly detect display column
			displayCol := getDisplayColumn(db, dbType, relatedTable)
			fmt.Printf("  -> Detected display column for %s: %s\n", relatedTable, displayCol)

			relations = append(relations, Relation{
				ForeignKey: c.Name,
				Table:      relatedTable,
				Alias:      baseName + "_" + displayCol, // e.g. user_username
				DisplayCol: displayCol,
			})
		}
	}
	fmt.Printf("Total relations found: %d\n", len(relations))

	// 4. Generate Artifacts
	// Model
	modelName := snakeToCamel(tableName)
	// Strip prefix
	camelPrefix := snakeToCamel(prefix)
	modelName = strings.TrimPrefix(modelName, camelPrefix)
	// Use singularize helper
	modelName = strings.Title(singularize(modelName))

	// Model
	createModel(modelName)

	// Auto-create related models
	for _, rel := range relations {
		relModelName := snakeToCamel(rel.Table)
		relModelName = strings.TrimPrefix(relModelName, camelPrefix)
		relModelName = strings.Title(singularize(relModelName))

		path := filepath.Join("app", "models", relModelName+".joss")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("Auto-creating missing related model: %s\n", relModelName)
			createModel(relModelName)
		}
	}

	// Controller
	createCRUDController(modelName, tableName, cols, relations)

	// Views
	if !isConsoleProject() {
		createCRUDViews(modelName, cols, relations)
		updateNavbar(modelName)
		injectProtectedRoutes(modelName)
	}
}

func createCRUDController(modelName, tableName string, cols []ColumnSchema, relations []Relation) {
	path := filepath.Join("app", "controllers", modelName+"Controller.joss")
	os.MkdirAll(filepath.Dir(path), 0755)

	viewPrefix := strings.ToLower(modelName)

	// Build Index Query with Joins
	// Note: GranDB now handles prefixes automatically for select() and joins()
	indexLogic := fmt.Sprintf(`$%s = new %s()`, strings.ToLower(modelName), modelName)

	if len(relations) > 0 {
		indexLogic += fmt.Sprintf("\n        $data = $%s", strings.ToLower(modelName))

		// Selects
		// Use base table names, ORM will prefix them
		_, _, _, _, _, _, currentPrefix := loadEnvConfig()
		baseTableName := strings.TrimPrefix(tableName, currentPrefix)

		selects := []string{fmt.Sprintf("\"%s.*\"", baseTableName)}

		for _, rel := range relations {
			baseRelTable := strings.TrimPrefix(rel.Table, currentPrefix)
			selects = append(selects, fmt.Sprintf("\"%s.%s as %s\"", baseRelTable, rel.DisplayCol, rel.Alias))
		}
		indexLogic += fmt.Sprintf(".select([%s])", strings.Join(selects, ", "))

		// Joins
		for _, rel := range relations {
			baseRelTable := strings.TrimPrefix(rel.Table, currentPrefix)
			indexLogic += fmt.Sprintf(".leftJoin(\"%s\", \"%s.%s\", \"=\", \"%s.id\")", baseRelTable, baseTableName, rel.ForeignKey, baseRelTable)
		}
		indexLogic += ".get()"
	} else {
		indexLogic += fmt.Sprintf("\n        $data = $%s.get()", strings.ToLower(modelName))
	}

	// Build Create Logic (Fetch relations)
	createLogic := ""
	createVars := ""
	if len(relations) > 0 {
		for _, rel := range relations {
			// Derive model name from table: js_roles -> Role
			relModel := snakeToCamel(rel.Table)
			// Get current prefix to strip
			_, _, _, _, _, _, prefix := loadEnvConfig()
			camelPrefix := snakeToCamel(prefix)
			relModel = strings.TrimPrefix(relModel, camelPrefix)
			relModel = strings.Title(singularize(relModel))
			varName := strings.ToLower(pluralize(relModel)) // roles
			createLogic += fmt.Sprintf("\n        $%sModel = new %s()", strings.ToLower(relModel), relModel)
			createLogic += fmt.Sprintf("\n        $%s = $%sModel.get()", varName, strings.ToLower(relModel))
			createVars += fmt.Sprintf(", \"%s\": $%s", varName, varName)
		}
	}

	content := fmt.Sprintf(`class %sController {
    
    function index() {
        %s
        return View.render("%s/index", {"items": $data})
    }

    function create() {
        %s
        return View.render("%s/create", {%s})
    }

    function store() {
        $req = new Request()
        $data = $req.except(["_token", "_referer", "_method"])
        
        $model = new %s()
        $model.insert($data)
        
        return redirect("/%s")
    }

    function edit($id) {
        $model = new %s()
        $item = $model.where("id", $id).first()
        %s
        return View.render("%s/edit", {"item": $item%s})
    }

    function update($id) {
        $req = new Request()
        $data = $req.except(["_token", "_referer", "_method"])
        
        $model = new %s()
        $model.where("id", $id).update($data)
        
        return redirect("/%s")
    }

    function delete($id) {
        $model = new %s()
        $model.where("id", $id).delete()
        return redirect("/%s")
    }
}`, modelName, indexLogic, viewPrefix, createLogic, viewPrefix, strings.TrimPrefix(createVars, ", "), modelName, viewPrefix, modelName, createLogic, viewPrefix, createVars, modelName, viewPrefix, modelName, viewPrefix)

	writeGenFile(path, content)
}

func createCRUDViews(modelName string, cols []ColumnSchema, relations []Relation) {
	folder := filepath.Join("app", "views", strings.ToLower(modelName))
	os.MkdirAll(folder, 0755)

	// Index
	indexHtml := fmt.Sprintf(`@extends('layouts.master')

@section('content')
<div class="card">
    <div class="card-header d-flex justify-content-between align-items-center">
        <h2>%s List</h2>
        <a href="/%s/create" class="btn btn-primary"><i class="fas fa-plus"></i> Create New</a>
    </div>
    <div class="card-body p-0">
        <div class="table-responsive">
            <table class="table table-hover mb-0">
                <thead>
                    <tr>
                        %s
                        <th class="text-end">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    @foreach($items as $item)
                    <tr>
                        %s
                        <td class="text-end">
                            <a href="/%s/edit/{{ $item.id }}" class="btn btn-sm btn-outline-info"><i class="fas fa-edit"></i></a>
                            <a href="/%s/delete/{{ $item.id }}" class="btn btn-sm btn-outline-danger" onclick="return confirm('Are you sure?')"><i class="fas fa-trash"></i></a>
                        </td>
                    </tr>
                    @endforeach
                </tbody>
            </table>
        </div>
    </div>
</div>
@endsection
`, modelName, strings.ToLower(modelName), generateIndexHeaders(cols, relations), generateIndexRows(cols, relations), strings.ToLower(modelName), strings.ToLower(modelName))

	writeGenFile(filepath.Join(folder, "index.joss.html"), indexHtml)

	// Create
	createHtml := fmt.Sprintf(`@extends('layouts.master')

@section('content')
<div class="row justify-content-center">
    <div class="col-md-8">
        <div class="card">
            <div class="card-header">
                <h2>Create %s</h2>
            </div>
            <div class="card-body">
                <form action="/%s/store" method="POST">
                    {{ csrf_field() }}
%s
                    <div class="d-flex justify-content-end gap-2">
                        <a href="/`+strings.ToLower(modelName)+`" class="btn btn-outline-secondary">Cancel</a>
                        <button type="submit" class="btn btn-primary">Save Record</button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</div>
@endsection
`, modelName, strings.ToLower(modelName), generateFormFields(cols, relations, false))
	writeGenFile(filepath.Join(folder, "create.joss.html"), createHtml)

	// Edit
	editHtml := fmt.Sprintf(`@extends('layouts.master')

@section('content')
<div class="row justify-content-center">
    <div class="col-md-8">
        <div class="card">
            <div class="card-header">
                <h2>Edit %s</h2>
            </div>
            <div class="card-body">
                <form action="/%s/update/{{ $item.id }}" method="POST">
                    {{ csrf_field() }}
%s
                    <div class="d-flex justify-content-end gap-2">
                        <a href="/`+strings.ToLower(modelName)+`" class="btn btn-outline-secondary">Cancel</a>
                        <button type="submit" class="btn btn-primary">Update Record</button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</div>
@endsection
`, modelName, strings.ToLower(modelName), generateFormFields(cols, relations, true))
	writeGenFile(filepath.Join(folder, "edit.joss.html"), editHtml)
}

// Helpers for view generation to keep createCRUDViews somewhat clean
func generateIndexHeaders(cols []ColumnSchema, relations []Relation) string {
	var html string
	for _, c := range cols {
		headerName := c.Name
		for _, rel := range relations {
			if c.Name == rel.ForeignKey {
				headerName = strings.Title(strings.Replace(rel.Alias, "_", " ", -1))
				break
			}
		}
		html += fmt.Sprintf("                        <th>%s</th>\n", headerName)
	}
	return html
}

func generateIndexRows(cols []ColumnSchema, relations []Relation) string {
	var html string
	for _, c := range cols {
		val := fmt.Sprintf("{{ $item.%s }}", c.Name)
		for _, rel := range relations {
			if c.Name == rel.ForeignKey {
				val = fmt.Sprintf("{{ $item.%s }}", rel.Alias)
				break
			}
		}
		html += fmt.Sprintf("                        <td>%s</td>\n", val)
	}
	return html
}

func generateFormFields(cols []ColumnSchema, relations []Relation, isEdit bool) string {
	var html string

	for _, c := range cols {
		if c.Name == "id" || c.Name == "created_at" || c.Name == "updated_at" {
			continue
		}

		// Check for relation
		isRelation := false
		var relData Relation
		for _, rel := range relations {
			if c.Name == rel.ForeignKey {
				isRelation = true
				relData = rel
				break
			}
		}

		if isRelation {
			// Derive variable name: js_roles -> roles
			relModel := snakeToCamel(relData.Table)
			_, _, _, _, _, _, prefix := loadEnvConfig()
			camelPrefix := snakeToCamel(prefix)
			relModel = strings.TrimPrefix(relModel, camelPrefix)
			relModel = strings.Title(singularize(relModel))
			varName := strings.ToLower(pluralize(relModel))

			selectedValue := ""
			if isEdit {
				selectedValue = fmt.Sprintf("{{ $item.%s == $opt.id ? 'selected' : '' }}", c.Name)
			}

			html += fmt.Sprintf(`                    <div class="form-group">
                        <label>%s</label>
                        <select name="%s" class="form-control">
                            <option value="">Select %s</option>
                            @foreach($%s as $opt)
                            <option value="{{ $opt.id }}" %s>{{ $opt.%s }}</option>
                            @endforeach
                        </select>
                    </div>
`, strings.Title(strings.Replace(c.Name, "_", " ", -1)), c.Name, relModel, varName, selectedValue, relData.DisplayCol)
		} else {
			valueAttr := ""
			if isEdit {
				valueAttr = fmt.Sprintf(" value=\"{{ $item.%s }}\"", c.Name)
			}
			html += fmt.Sprintf(`                    <div class="form-group">
                        <label>%s</label>
                        <input type="text" name="%s" class="form-control"%s>
                    </div>
`, c.Name, c.Name, valueAttr)
		}
	}
	return html
}

func updateNavbar(modelName string) {
	// Try to find layouts/master.joss.html
	path := filepath.Join("app", "views", "layouts", "master.joss.html")
	content, err := ioutil.ReadFile(path)
	if err == nil {
		html := string(content)
		link := fmt.Sprintf(`<li><a href="/%s"><i class="fas fa-circle"></i> %s</a></li>`, strings.ToLower(modelName), modelName)

		// Insert before <!-- Injected Links Here -->
		if strings.Contains(html, "<!-- Injected Links Here -->") {
			html = strings.Replace(html, "<!-- Injected Links Here -->", link+"\n                        <!-- Injected Links Here -->", 1)
			ioutil.WriteFile(path, []byte(html), 0644)
			fmt.Println("Updated navbar in layouts/master.joss.html")
		}
	}
}

func injectProtectedRoutes(modelName string) {
	path := "routes.joss"
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	routes := fmt.Sprintf(`
    // CRUD Routes for %s
    Router.get("/%s", "%sController@index")
    Router.get("/%s/create", "%sController@create")
    Router.post("/%s/store", "%sController@store")
    Router.get("/%s/edit/{id}", "%sController@edit")
    Router.post("/%s/update/{id}", "%sController@update")
    Router.get("/%s/delete/{id}", "%sController@delete")
`, modelName,
		strings.ToLower(modelName), modelName,
		strings.ToLower(modelName), modelName,
		strings.ToLower(modelName), modelName,
		strings.ToLower(modelName), modelName,
		strings.ToLower(modelName), modelName,
		strings.ToLower(modelName), modelName)

	strContent := string(content)

	// Check if "auth" group exists
	if strings.Contains(strContent, `Router.group("auth"`) {
		// Inject inside group
		// Simple approach: Find `Router.group("auth", function() {` and append after it.
		target := `Router.group("auth", function() {`
		if strings.Contains(strContent, target) {
			strContent = strings.Replace(strContent, target, target+routes, 1)
			ioutil.WriteFile(path, []byte(strContent), 0644)
			fmt.Println("Injected protected routes into 'auth' group.")
			return
		}
	}

	// If no group, append a new protected group
	newGroup := fmt.Sprintf(`
Router.group("auth", function() {%s})
`, routes)

	ioutil.WriteFile(path, []byte(strContent+newGroup), 0644)
	fmt.Println("Created new 'auth' group with routes.")
}

func removeCRUD(tableName string) {
	fmt.Printf("Removing CRUD for table '%s'...\n", tableName)

	// 1. Infer Model Name
	modelName := snakeToCamel(tableName)
	// Strip prefix
	_, _, _, _, _, _, prefix := loadEnvConfig()
	camelPrefix := snakeToCamel(prefix)
	modelName = strings.TrimPrefix(modelName, camelPrefix)
	modelName = strings.Title(singularize(modelName))

	fmt.Printf("Inferred Model Name: %s\n", modelName)

	// 2. Delete Controller
	controllerPath := filepath.Join("app", "controllers", modelName+"Controller.joss")
	if _, err := os.Stat(controllerPath); err == nil {
		os.Remove(controllerPath)
		fmt.Printf("Deleted: %s\n", controllerPath)
	}

	// 3. Delete Model
	modelPath := filepath.Join("app", "models", modelName+".joss")
	if _, err := os.Stat(modelPath); err == nil {
		os.Remove(modelPath)
		fmt.Printf("Deleted: %s\n", modelPath)
	}

	// 4. Delete Views
	viewsPath := filepath.Join("app", "views", strings.ToLower(modelName))
	if _, err := os.Stat(viewsPath); err == nil {
		os.RemoveAll(viewsPath)
		fmt.Printf("Deleted: %s\n", viewsPath)
	}

	// 5. Remove Routes
	routesPath := "routes.joss"
	content, err := ioutil.ReadFile(routesPath)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		var newLines []string
		for _, line := range lines {
			// Check for CRUD Routes comment
			if strings.Contains(line, fmt.Sprintf("// CRUD Routes for %s", modelName)) {
				continue
			}
			// Filter out lines that contain the controller name (case insensitive check)
			lowerLine := strings.ToLower(line)
			lowerController := strings.ToLower(modelName + "Controller")
			if strings.Contains(lowerLine, lowerController) {
				continue
			}
			newLines = append(newLines, line)
		}
		ioutil.WriteFile(routesPath, []byte(strings.Join(newLines, "\n")), 0644)
		fmt.Println("Cleaned routes.")
	}

	// 6. Remove Navbar Link
	masterPath := filepath.Join("app", "views", "layouts", "master.joss.html")
	content, err = ioutil.ReadFile(masterPath)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		var newLines []string
		linkTarget := fmt.Sprintf(`href="/%s"`, strings.ToLower(modelName))
		for _, line := range lines {
			if strings.Contains(line, linkTarget) {
				continue
			}
			newLines = append(newLines, line)
		}
		ioutil.WriteFile(masterPath, []byte(strings.Join(newLines, "\n")), 0644)
		fmt.Println("Cleaned navbar.")
	}

	fmt.Println("CRUD removal complete.")
}
