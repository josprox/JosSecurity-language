package analyzer

import (
	"fmt"
	"strings"

	"github.com/jossecurity/joss/pkg/diagnostics"
	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/typesystem"
)

func (a *Analyzer) inferExpression(expression parser.Expression, current *scope) typesystem.Type {
	if expression == nil {
		return typesystem.Type{Kind: typesystem.Unknown}
	}
	switch node := expression.(type) {
	case *parser.StringLiteral:
		return typesystem.Type{Kind: typesystem.String}
	case *parser.IntegerLiteral:
		return typesystem.Type{Kind: typesystem.Int}
	case *parser.FloatLiteral:
		return typesystem.Type{Kind: typesystem.Float}
	case *parser.Boolean:
		return typesystem.Type{Kind: typesystem.Bool}
	case *parser.NullLiteral:
		return typesystem.Type{Kind: typesystem.Null}
	case *parser.Identifier:
		return a.inferIdentifier(node, current)
	case *parser.ArrayLiteral:
		for _, element := range node.Elements {
			a.inferExpression(element, current)
		}
		return typesystem.Type{Kind: typesystem.Array}
	case *parser.MapLiteral:
		for key, value := range node.Pairs {
			keyType := a.inferExpression(key, current)
			if keyType.IsKnown() && keyType.Kind != typesystem.String {
				a.add("JOSS-TYPE-005", diagnostics.SeverityError, a.file, tokenOfExpression(key),
					fmt.Sprintf("Map key has type `%s`; Joss maps require `string` keys.", keyType.String()),
					"The runtime indexes maps by string.", "Convert the key to string.")
			}
			a.inferExpression(value, current)
		}
		return typesystem.Type{Kind: typesystem.Map}
	case *parser.AssignExpression:
		return a.inferAssignment(node, current)
	case *parser.InfixExpression:
		return a.inferInfix(node, current)
	case *parser.PrefixExpression:
		valueType := a.inferExpression(node.Right, current)
		if node.Operator == "!" {
			return typesystem.Type{Kind: typesystem.Bool}
		}
		if node.Operator == "-" && valueType.IsKnown() && !isNumeric(valueType) {
			a.invalidOperator(node.Token, node.Operator, valueType, typesystem.Type{})
		}
		return valueType
	case *parser.PostfixExpression:
		valueType := a.inferExpression(node.Left, current)
		if valueType.IsKnown() && !isNumeric(valueType) {
			a.invalidOperator(node.Token, node.Operator, valueType, typesystem.Type{})
		}
		return valueType
	case *parser.TernaryExpression:
		a.inferExpression(node.Condition, current)
		trueType := a.inferExpression(node.True, current)
		if node.True == nil {
			trueType = a.inferExpression(node.Condition, current)
		}
		falseType := a.inferExpression(node.False, current)
		return commonType(trueType, falseType)
	case *parser.IndexExpression:
		containerType := a.inferExpression(node.Left, current)
		indexType := a.inferExpression(node.Index, current)
		if containerType.IsKnown() {
			switch containerType.Kind {
			case typesystem.Array, typesystem.String:
				if indexType.IsKnown() && indexType.Kind != typesystem.Int {
					a.add("JOSS-TYPE-006", diagnostics.SeverityError, a.file, node.Token,
						fmt.Sprintf("`%s` values require an `int` index, got `%s`.", containerType.String(), indexType.String()),
						"Index type is known before execution.", "Use an integer index.")
				}
			case typesystem.Map:
				if indexType.IsKnown() && indexType.Kind != typesystem.String {
					a.add("JOSS-TYPE-006", diagnostics.SeverityError, a.file, node.Token,
						fmt.Sprintf("`map` values require a `string` index, got `%s`.", indexType.String()),
						"Map keys are strings in the Joss runtime.", "Use a string key.")
				}
			default:
				a.add("JOSS-TYPE-007", diagnostics.SeverityError, a.file, node.Token,
					fmt.Sprintf("Values of type `%s` cannot be indexed.", containerType.String()),
					"Only arrays, maps and strings support index access.", "Check the value or use member access.")
			}
		}
		if containerType.Kind == typesystem.String {
			return typesystem.Type{Kind: typesystem.String}
		}
		return typesystem.Type{Kind: typesystem.Unknown}
	case *parser.NewExpression:
		return a.inferNew(node, current)
	case *parser.MemberExpression:
		return a.inferMember(node, current)
	case *parser.CallExpression:
		return a.inferCall(node, current)
	case *parser.FunctionLiteral:
		a.analyzeCallable(node.Parameters, node.Body, current, "")
		return typesystem.Type{Kind: typesystem.Object}
	case *parser.IssetExpression:
		a.suppressUndefined++
		for _, argument := range node.Arguments {
			a.inferExpression(argument, current)
		}
		a.suppressUndefined--
		return typesystem.Type{Kind: typesystem.Bool}
	case *parser.EmptyExpression:
		a.suppressUndefined++
		a.inferExpression(node.Argument, current)
		a.suppressUndefined--
		return typesystem.Type{Kind: typesystem.Bool}
	case *parser.BlockExpression:
		if node.Block != nil {
			a.analyzeBlock(node.Block, current)
		}
		return typesystem.Type{Kind: typesystem.Unknown}
	case *parser.MatchExpression:
		a.inferExpression(node.Subject, current)
		result := typesystem.Type{Kind: typesystem.Unknown}
		for _, arm := range node.Arms {
			if !arm.IsDefault {
				for _, key := range arm.Keys {
					if identifier, ok := key.(*parser.Identifier); ok && identifier.Value == "default" {
						continue
					}
					a.inferExpression(key, current)
				}
			}
			result = commonType(result, a.inferExpression(arm.Value, current))
		}
		return result
	default:
		return typesystem.Type{Kind: typesystem.Unknown}
	}
}

func (a *Analyzer) inferIdentifier(identifier *parser.Identifier, current *scope) typesystem.Type {
	name := cleanName(identifier.Value)
	switch name {
	case "null", "nil":
		return typesystem.Type{Kind: typesystem.Null}
	case "true", "false":
		return typesystem.Type{Kind: typesystem.Bool}
	case "self", "this":
		if a.currentClass != "" {
			return typesystem.Type{Kind: typesystem.Class, Name: a.currentClass}
		}
	case "parent", "super":
		if class, exists := a.classes[a.currentClass]; exists && class.SuperClass != "" {
			return typesystem.Type{Kind: typesystem.Class, Name: class.SuperClass}
		}
	}
	if value, exists := current.resolve(name); exists {
		value.Used = true
		return value.Type
	}
	if _, exists := a.classes[name]; exists {
		return typesystem.Type{Kind: typesystem.Class, Name: name}
	}
	if _, exists := a.functions[name]; exists {
		return typesystem.Type{Kind: typesystem.Object}
	}
	if _, exists := a.environment.Builtins[name]; exists {
		return typesystem.Type{Kind: typesystem.Object}
	}
	// Constants are resolved by runtime/environment integrations. Until Joss has
	// an explicit const declaration node, an all-uppercase identifier remains
	// unknown rather than being misreported as a variable.
	if name != "" && strings.ToUpper(name) == name && len(name) > 1 {
		return typesystem.Type{Kind: typesystem.Unknown}
	}
	if a.suppressUndefined == 0 {
		a.add("JOSS-SYM-001", diagnostics.SeverityError, a.file, identifier.Token,
			fmt.Sprintf("Variable `$%s` is used before it is declared.", name),
			"The active lexical scope has no symbol with this name.", "Declare or assign the variable before this use.")
	}
	return typesystem.Type{Kind: typesystem.Unknown}
}

func (a *Analyzer) inferAssignment(assignment *parser.AssignExpression, current *scope) typesystem.Type {
	valueType := a.inferExpression(assignment.Value, current)
	if identifier, ok := assignment.Left.(*parser.Identifier); ok {
		name := cleanName(identifier.Value)
		if existing, exists := current.resolve(name); exists {
			if existing.Inferred && !existing.Type.IsKnown() {
				existing.Type = typesystem.MergeInference(existing.Type, valueType)
			}
			if !existing.Dynamic && !assignableExpression(existing.Type, valueType, assignment.Value) {
				a.typeMismatch("JOSS-TYPE-001", name, existing.Type, valueType, identifier.Token, "assignment")
			}
			return existing.Type
		}
		inferredType := typesystem.MergeInference(typesystem.Type{Kind: typesystem.Unknown}, valueType)
		current.put(&symbol{Name: name, Type: inferredType, Kind: symbolVariable, Token: identifier.Token, File: a.file, Inferred: true})
		return inferredType
	}
	// Member/index assignment still needs the receiver and index checked.
	a.inferExpression(assignment.Left, current)
	return valueType
}

func (a *Analyzer) inferInfix(expression *parser.InfixExpression, current *scope) typesystem.Type {
	left := a.inferExpression(expression.Left, current)
	right := a.inferExpression(expression.Right, current)
	switch expression.Operator {
	case ".":
		return typesystem.Type{Kind: typesystem.String}
	case "==", "!=", "===", "!==", "<", ">", "<=", ">=", "&&", "||":
		return typesystem.Type{Kind: typesystem.Bool}
	case "<=>":
		return typesystem.Type{Kind: typesystem.Int}
	case "??":
		return commonType(left, right)
	case "+", "-", "*", "/", "%":
		if left.IsKnown() && !isNumeric(left) {
			a.invalidOperator(expression.Token, expression.Operator, left, right)
			return typesystem.Type{Kind: typesystem.Unknown}
		}
		if right.IsKnown() && !isNumeric(right) {
			a.invalidOperator(expression.Token, expression.Operator, left, right)
			return typesystem.Type{Kind: typesystem.Unknown}
		}
		if expression.Operator == "/" || left.Kind == typesystem.Float || right.Kind == typesystem.Float {
			return typesystem.Type{Kind: typesystem.Float}
		}
		if left.Kind == typesystem.Int && right.Kind == typesystem.Int {
			return typesystem.Type{Kind: typesystem.Int}
		}
		return typesystem.Type{Kind: typesystem.Unknown}
	case "<<", ">>", "|>":
		return typesystem.Type{Kind: typesystem.Unknown}
	default:
		return typesystem.Type{Kind: typesystem.Unknown}
	}
}

func (a *Analyzer) invalidOperator(token parser.Token, operator string, left, right typesystem.Type) {
	message := fmt.Sprintf("Operator `%s` is not defined for `%s`.", operator, left.String())
	if right.Kind != "" {
		message = fmt.Sprintf("Operator `%s` is not defined for `%s` and `%s`.", operator, left.String(), right.String())
	}
	a.add("JOSS-TYPE-004", diagnostics.SeverityError, a.file, token, message,
		"The operand types are known and incompatible with this operator.", "Convert the operands or use an operator defined for these types.")
}

func (a *Analyzer) inferNew(expression *parser.NewExpression, current *scope) typesystem.Type {
	for _, argument := range expression.Arguments {
		a.inferExpression(argument, current)
	}
	if expression.Class == nil {
		return typesystem.Type{Kind: typesystem.Unknown}
	}
	className := expression.Class.Value
	class, exists := a.classes[className]
	if !exists {
		a.add("JOSS-SYM-004", diagnostics.SeverityError, a.file, expression.Class.Token,
			fmt.Sprintf("Class `%s` does not exist.", className),
			"The class was not found in project sources, native classes or loaded plugin symbols.", "Check the class name and plugin configuration.")
		return typesystem.Type{Kind: typesystem.Unknown}
	}
	if constructor, ok := class.Methods["constructor"]; ok {
		a.checkCall(constructor, expression.Arguments, current, expression.Class.Token)
	}
	return typesystem.Type{Kind: typesystem.Class, Name: className}
}

func (a *Analyzer) inferMember(expression *parser.MemberExpression, current *scope) typesystem.Type {
	if expression == nil {
		return typesystem.Type{Kind: typesystem.Unknown}
	}
	return a.receiverType(expression.Left, current)
}

func (a *Analyzer) inferCall(call *parser.CallExpression, current *scope) typesystem.Type {
	if identifier, ok := call.Function.(*parser.Identifier); ok {
		name := identifier.Value
		if builtin, exists := a.environment.Builtins[name]; exists {
			a.checkCall(builtin, call.Arguments, current, identifier.Token)
			return builtin.ReturnType
		}
		if function, exists := a.functions[name]; exists {
			a.checkCall(function.callable, call.Arguments, current, identifier.Token)
			return function.callable.ReturnType
		}
		if variable, exists := current.resolve(name); exists {
			variable.Used = true
			for _, argument := range call.Arguments {
				a.inferExpression(argument, current)
			}
			return typesystem.Type{Kind: typesystem.Unknown}
		}
		for _, argument := range call.Arguments {
			a.inferExpression(argument, current)
		}
		a.add("JOSS-SYM-003", diagnostics.SeverityError, a.file, identifier.Token,
			fmt.Sprintf("Function `%s` does not exist.", name),
			"No project function, runtime builtin or callable variable has this name.", "Declare the function or check its spelling.")
		return typesystem.Type{Kind: typesystem.Unknown}
	}
	if member, ok := call.Function.(*parser.MemberExpression); ok {
		receiver := a.receiverType(member.Left, current)
		if receiver.Kind == typesystem.Class && member.Property != nil {
			if callable, exists := a.lookupMethod(receiver.Name, member.Property.Value); exists {
				a.checkCall(callable, call.Arguments, current, member.Property.Token)
				return callable.ReturnType
			}
			a.add("JOSS-MEMBER-001", diagnostics.SeverityError, a.file, member.Property.Token,
				fmt.Sprintf("Class `%s` has no method `%s`.", receiver.Name, member.Property.Value),
				"The receiver class is known and its method table has been resolved.", "Check the method name or the class API.")
		}
		for _, argument := range call.Arguments {
			a.inferExpression(argument, current)
		}
		return typesystem.Type{Kind: typesystem.Unknown}
	}
	functionType := a.inferExpression(call.Function, current)
	for _, argument := range call.Arguments {
		a.inferExpression(argument, current)
	}
	return functionType
}

func (a *Analyzer) receiverType(expression parser.Expression, current *scope) typesystem.Type {
	if identifier, ok := expression.(*parser.Identifier); ok {
		// A local may intentionally have the same spelling as its class (a common
		// model pattern). Lexical symbols shadow class names for instance access.
		if value, exists := current.resolve(identifier.Value); exists {
			value.Used = true
			return value.Type
		}
		if _, exists := a.classes[identifier.Value]; exists {
			return typesystem.Type{Kind: typesystem.Class, Name: identifier.Value}
		}
	}
	return a.inferExpression(expression, current)
}

func (a *Analyzer) lookupMethod(className, methodName string) (Callable, bool) {
	visited := map[string]bool{}
	for className != "" && !visited[className] {
		visited[className] = true
		class, exists := a.classes[className]
		if !exists {
			return Callable{}, false
		}
		if method, exists := class.Methods[methodName]; exists {
			return method, true
		}
		className = class.SuperClass
	}
	return Callable{}, false
}

func (a *Analyzer) checkCall(callable Callable, arguments []parser.Expression, current *scope, token parser.Token) {
	argumentTypes := make([]typesystem.Type, len(arguments))
	for index, argument := range arguments {
		argumentTypes[index] = a.inferExpression(argument, current)
	}
	minimum := 0
	for _, parameter := range callable.Parameters {
		if !parameter.HasDefault {
			minimum++
		}
	}
	if len(arguments) < minimum || (!callable.Variadic && len(arguments) > len(callable.Parameters)) {
		expected := fmt.Sprintf("%d", len(callable.Parameters))
		if minimum != len(callable.Parameters) {
			expected = fmt.Sprintf("%d..%d", minimum, len(callable.Parameters))
		}
		if callable.Variadic {
			expected = fmt.Sprintf("at least %d", minimum)
		}
		a.add("JOSS-CALL-001", diagnostics.SeverityError, a.file, token,
			fmt.Sprintf("Call to `%s` expects %s argument(s), got %d.", callable.Name, expected, len(arguments)),
			"The callable signature is known at analysis time.", "Pass the required arguments or update the function signature.")
	}
	for index := 0; index < len(arguments) && index < len(callable.Parameters); index++ {
		parameter := callable.Parameters[index]
		if !assignableExpression(parameter.Type, argumentTypes[index], arguments[index]) {
			a.add("JOSS-TYPE-003", diagnostics.SeverityError, a.file, tokenOfExpression(arguments[index]),
				fmt.Sprintf("Argument %d to `%s` has type `%s`; parameter `$%s` requires `%s`.", index+1, callable.Name, argumentTypes[index].String(), parameter.Name, parameter.Type.String()),
				"Function arguments follow the same assignment compatibility rules as variables.", "Convert the argument or correct the parameter type.")
		}
	}
}

func isNumeric(valueType typesystem.Type) bool {
	return valueType.Kind == typesystem.Int || valueType.Kind == typesystem.Float
}

func commonType(left, right typesystem.Type) typesystem.Type {
	if left.Kind == right.Kind && (left.Kind != typesystem.Class || left.Name == right.Name) {
		return left
	}
	if !left.IsKnown() {
		return right
	}
	if !right.IsKnown() {
		return left
	}
	if isNumeric(left) && isNumeric(right) {
		return typesystem.Type{Kind: typesystem.Float}
	}
	return typesystem.Type{Kind: typesystem.Unknown}
}

func assignableExpression(destination, source typesystem.Type, expression parser.Expression) bool {
	if typesystem.Assignable(destination, source) {
		return true
	}
	if literal, ok := expression.(*parser.StringLiteral); ok {
		_, coerced := typesystem.CoerceString(destination, literal.Value)
		return coerced
	}
	return false
}

func tokenOfExpression(expression parser.Expression) parser.Token {
	switch node := expression.(type) {
	case *parser.Identifier:
		return node.Token
	case *parser.StringLiteral:
		return node.Token
	case *parser.IntegerLiteral:
		return node.Token
	case *parser.FloatLiteral:
		return node.Token
	case *parser.Boolean:
		return node.Token
	case *parser.NullLiteral:
		return node.Token
	case *parser.ArrayLiteral:
		return node.Token
	case *parser.MapLiteral:
		return node.Token
	case *parser.AssignExpression:
		return node.Token
	case *parser.InfixExpression:
		return node.Token
	case *parser.PrefixExpression:
		return node.Token
	case *parser.PostfixExpression:
		return node.Token
	case *parser.TernaryExpression:
		return node.Token
	case *parser.IndexExpression:
		return node.Token
	case *parser.NewExpression:
		return node.Token
	case *parser.MemberExpression:
		return node.Token
	case *parser.CallExpression:
		return node.Token
	case *parser.FunctionLiteral:
		return node.Token
	case *parser.IssetExpression:
		return node.Token
	case *parser.EmptyExpression:
		return node.Token
	case *parser.BlockExpression:
		return node.Token
	case *parser.MatchExpression:
		return node.Token
	default:
		return parser.Token{}
	}
}
