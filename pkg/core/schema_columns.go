package core

import (
	"fmt"
	"strings"
)

func (r *Runtime) buildColumnDefinition(name, typeStr, driver string) string {
	parts := strings.Split(typeStr, "|")
	baseType := parts[0]
	modifiers := parts[1:]

	var sqlDef string

	// Parse base type (handle arguments like char(100))
	typeName := baseType
	typeArgs := ""
	if strings.Contains(baseType, "(") {
		start := strings.Index(baseType, "(")
		end := strings.LastIndex(baseType, ")")
		typeName = baseType[:start]
		typeArgs = baseType[start+1 : end]
	}

	switch typeName {
	case "increments":
		if driver == "sqlite" {
			sqlDef = "INTEGER PRIMARY KEY AUTOINCREMENT"
		} else if driver == "postgres" {
			sqlDef = "SERIAL PRIMARY KEY"
		} else {
			sqlDef = "INT AUTO_INCREMENT PRIMARY KEY"
		}
	case "bigIncrements":
		if driver == "sqlite" {
			sqlDef = "INTEGER PRIMARY KEY AUTOINCREMENT"
		} else if driver == "postgres" {
			sqlDef = "BIGSERIAL PRIMARY KEY"
		} else {
			sqlDef = "BIGINT AUTO_INCREMENT PRIMARY KEY"
		}
	case "tinyInteger":
		if driver == "postgres" {
			sqlDef = "SMALLINT"
		} else {
			sqlDef = "TINYINT"
		}
	case "smallInteger":
		sqlDef = "SMALLINT"
	case "mediumInteger":
		if driver == "postgres" {
			sqlDef = "INTEGER"
		} else {
			sqlDef = "MEDIUMINT"
		}
	case "integer":
		sqlDef = "INTEGER"
	case "bigInteger":
		sqlDef = "BIGINT"
	case "float":
		if driver == "postgres" {
			sqlDef = "REAL"
		} else {
			sqlDef = "FLOAT"
		}
	case "double":
		if driver == "postgres" {
			sqlDef = "DOUBLE PRECISION"
		} else {
			sqlDef = "DOUBLE"
		}
	case "decimal":
		precision := "8,2"
		if typeArgs != "" {
			precision = typeArgs
		}
		sqlDef = fmt.Sprintf("DECIMAL(%s)", precision)
	case "char":
		length := "255"
		if typeArgs != "" {
			length = typeArgs
		}
		sqlDef = fmt.Sprintf("CHAR(%s)", length)
	case "string":
		length := "255"
		if typeArgs != "" {
			length = typeArgs
		}
		sqlDef = fmt.Sprintf("VARCHAR(%s)", length)
	case "text":
		sqlDef = "TEXT"
	case "mediumText":
		if driver == "sqlite" || driver == "postgres" {
			sqlDef = "TEXT"
		} else {
			sqlDef = "MEDIUMTEXT"
		}
	case "longText":
		if driver == "sqlite" || driver == "postgres" {
			sqlDef = "TEXT"
		} else {
			sqlDef = "LONGTEXT"
		}
	case "date":
		sqlDef = "DATE"
	case "dateTime":
		if driver == "postgres" {
			sqlDef = "TIMESTAMP"
		} else {
			sqlDef = "DATETIME"
		}
	case "time":
		sqlDef = "TIME"
	case "timestamp":
		sqlDef = "TIMESTAMP"
	case "boolean":
		if driver == "sqlite" || driver == "postgres" {
			sqlDef = "BOOLEAN"
		} else {
			sqlDef = "TINYINT(1)"
		}
	case "json":
		if driver == "sqlite" {
			sqlDef = "TEXT"
		} else if driver == "postgres" {
			sqlDef = "JSONB"
		} else {
			sqlDef = "JSON"
		}
	case "enum":
		if driver == "sqlite" || driver == "postgres" {
			sqlDef = "TEXT"
		} else {
			sqlDef = fmt.Sprintf("ENUM(%s)", typeArgs)
		}
	default:
		sqlDef = "VARCHAR(255)"
	}

	def := fmt.Sprintf("%s %s", name, sqlDef)

	// Process modifiers
	isUnsigned := false
	isNullable := false

	for _, mod := range modifiers {
		if mod == "unsigned" {
			isUnsigned = true
		} else if mod == "nullable" {
			isNullable = true
		}
	}

	// Apply Unsigned (MySQL only mostly)
	if isUnsigned && driver == "mysql" {
		if strings.Contains(strings.ToLower(sqlDef), "int") || strings.Contains(strings.ToLower(sqlDef), "double") || strings.Contains(strings.ToLower(sqlDef), "float") || strings.Contains(strings.ToLower(sqlDef), "decimal") {
			def = strings.Replace(def, sqlDef, sqlDef+" UNSIGNED", 1)
		}
	}

	// Apply Nullable
	if !isNullable && !strings.Contains(typeName, "increments") {
		def += " NOT NULL"
	} else if isNullable {
		def += " NULL"
	}

	// Apply Default
	for _, mod := range modifiers {
		if strings.HasPrefix(mod, "default") {
			start := strings.Index(mod, "(")
			end := strings.LastIndex(mod, ")")
			if start != -1 && end != -1 {
				val := mod[start+1 : end]
				def += fmt.Sprintf(" DEFAULT %s", val)
			}
		}
	}

	// Apply Unique
	for _, mod := range modifiers {
		if mod == "unique" {
			def += " UNIQUE"
		}
	}

	// Apply Comment (MySQL only)
	if driver == "mysql" {
		for _, mod := range modifiers {
			if strings.HasPrefix(mod, "comment") {
				start := strings.Index(mod, "(")
				end := strings.LastIndex(mod, ")")
				if start != -1 && end != -1 {
					val := mod[start+1 : end]
					def += fmt.Sprintf(" COMMENT %s", val)
				}
			}
		}
	}

	return def
}
