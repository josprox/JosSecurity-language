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
		} else if driver == "sqlserver" {
			sqlDef = "INT IDENTITY(1,1) PRIMARY KEY"
		} else {
			sqlDef = "INT AUTO_INCREMENT PRIMARY KEY"
		}
	case "bigIncrements":
		if driver == "sqlite" {
			sqlDef = "INTEGER PRIMARY KEY AUTOINCREMENT"
		} else if driver == "postgres" {
			sqlDef = "BIGSERIAL PRIMARY KEY"
		} else if driver == "sqlserver" {
			sqlDef = "BIGINT IDENTITY(1,1) PRIMARY KEY"
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
		if driver == "postgres" || driver == "sqlserver" {
			sqlDef = "INTEGER"
		} else {
			sqlDef = "MEDIUMINT"
		}
	case "integer":
		if driver == "sqlserver" {
			sqlDef = "INT"
		} else {
			sqlDef = "INTEGER"
		}
	case "bigInteger":
		sqlDef = "BIGINT"
	case "float":
		if driver == "postgres" || driver == "sqlserver" {
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
		if driver == "sqlserver" {
			sqlDef = fmt.Sprintf("NCHAR(%s)", length)
		} else {
			sqlDef = fmt.Sprintf("CHAR(%s)", length)
		}
	case "string":
		length := "255"
		if typeArgs != "" {
			length = typeArgs
		}
		if driver == "sqlserver" {
			sqlDef = fmt.Sprintf("NVARCHAR(%s)", length)
		} else {
			sqlDef = fmt.Sprintf("VARCHAR(%s)", length)
		}
	case "text":
		if driver == "sqlserver" {
			sqlDef = "NVARCHAR(MAX)"
		} else {
			sqlDef = "TEXT"
		}
	case "mediumText":
		if driver == "sqlite" || driver == "postgres" {
			sqlDef = "TEXT"
		} else if driver == "sqlserver" {
			sqlDef = "NVARCHAR(MAX)"
		} else {
			sqlDef = "MEDIUMTEXT"
		}
	case "longText":
		if driver == "sqlite" || driver == "postgres" {
			sqlDef = "TEXT"
		} else if driver == "sqlserver" {
			sqlDef = "NVARCHAR(MAX)"
		} else {
			sqlDef = "LONGTEXT"
		}
	case "date":
		sqlDef = "DATE"
	case "dateTime":
		if driver == "postgres" {
			sqlDef = "TIMESTAMP"
		} else if driver == "sqlserver" {
			sqlDef = "DATETIME2"
		} else {
			sqlDef = "DATETIME"
		}
	case "time":
		sqlDef = "TIME"
	case "timestamp":
		if driver == "sqlserver" {
			sqlDef = "DATETIME2"
		} else {
			sqlDef = "TIMESTAMP"
		}
	case "boolean":
		if driver == "sqlite" || driver == "postgres" {
			sqlDef = "BOOLEAN"
		} else if driver == "sqlserver" {
			sqlDef = "BIT"
		} else {
			sqlDef = "TINYINT(1)"
		}
	case "json":
		if driver == "sqlite" {
			sqlDef = "TEXT"
		} else if driver == "postgres" {
			sqlDef = "JSONB"
		} else if driver == "sqlserver" {
			sqlDef = "NVARCHAR(MAX)"
		} else {
			sqlDef = "JSON"
		}
	case "enum":
		if driver == "sqlite" || driver == "postgres" {
			sqlDef = "TEXT"
		} else if driver == "sqlserver" {
			sqlDef = "NVARCHAR(255)"
		} else {
			sqlDef = fmt.Sprintf("ENUM(%s)", typeArgs)
		}
	default:
		if driver == "sqlserver" {
			sqlDef = "NVARCHAR(255)"
		} else {
			sqlDef = "VARCHAR(255)"
		}
	}

	def := fmt.Sprintf("%s %s", name, sqlDef)

	isUnsigned := false
	isNullable := false

	for _, mod := range modifiers {
		if mod == "unsigned" {
			isUnsigned = true
		} else if mod == "nullable" {
			isNullable = true
		}
	}

	if isUnsigned && driver == "mysql" {
		if strings.Contains(strings.ToLower(sqlDef), "int") || strings.Contains(strings.ToLower(sqlDef), "double") || strings.Contains(strings.ToLower(sqlDef), "float") || strings.Contains(strings.ToLower(sqlDef), "decimal") {
			def = strings.Replace(def, sqlDef, sqlDef+" UNSIGNED", 1)
		}
	}

	if !isNullable && !strings.Contains(typeName, "increments") {
		def += " NOT NULL"
	} else if isNullable {
		def += " NULL"
	}

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

	for _, mod := range modifiers {
		if mod == "unique" {
			def += " UNIQUE"
		}
	}

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
