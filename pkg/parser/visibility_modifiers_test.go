package parser

import (
	"testing"
)

func TestVisibilityAndStaticModifiers(t *testing.T) {
	sources := []string{
		`
		public class AuthService {
			private string $apiKey = "secret_123"
			public static $instance = null
			protected $attempts = 0

			public static func getInstance() {
				return AuthService::$instance
			}

			private func hashToken(string $token) {
				return md5($token)
			}

			protected func handleAuth(mixed $user, mixed $code = 200) {
				return json($user, $code)
			}
		}
		`,
		`
		public class BaseController {
			protected $db
			public func constructor() {
				$this->db = new GranDB()
			}
		}

		public class UserController extends BaseController {
			public func index(mixed $page = 1, mixed $limit = 10) {
				return $this->db->table("users")->paginate($limit, $page)
			}
		}
		`,
		`
		public func calcularTotal(float $precio, int $cantidad = 1, string $moneda = "USD") {
			return $precio * $cantidad
		}
		`,
	}

	for i, src := range sources {
		p := NewParser(NewLexer(src))
		prog := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("Source %d failed to parse: %v", i+1, p.Errors())
		}
		if len(prog.Statements) == 0 {
			t.Fatalf("Source %d resulted in 0 statements", i+1)
		}
	}
}

func TestFunctionDefaultParametersAST(t *testing.T) {
	src := `
	public func makeUser(string $name, int $roleId = 2, bool $active = true) {
		return { "name": $name, "role_id": $roleId, "active": $active }
	}
	`
	p := NewParser(NewLexer(src))
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Failed to parse default parameters: %v", p.Errors())
	}

	if len(prog.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(prog.Statements))
	}

	fn, ok := prog.Statements[0].(*MethodStatement)
	if !ok {
		t.Fatalf("Expected *MethodStatement, got %T", prog.Statements[0])
	}

	if len(fn.Parameters) != 3 {
		t.Fatalf("Expected 3 parameters, got %d", len(fn.Parameters))
	}

	if fn.Parameters[0].DefaultValue != nil {
		t.Errorf("Parameter 0 should not have default value")
	}

	if fn.Parameters[1].DefaultValue == nil {
		t.Errorf("Parameter 1 ($roleId) should have default value (2)")
	}

	if fn.Parameters[2].DefaultValue == nil {
		t.Errorf("Parameter 2 ($active) should have default value (true)")
	}
}
