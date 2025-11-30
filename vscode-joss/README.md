# JosSecurity VS Code Extension v2.0

Advanced language support for JosSecurity with Language Server Protocol (LSP).

## Features

### 🚀 Language Server Protocol (LSP)
- Full LSP implementation with TypeScript
- Real-time indexing of workspace
- Incremental updates on file changes

### 🔍 Navigation
- **Go-to-Definition** (Ctrl+Click / F12) for:
  - Controllers (`AuthController`)
  - Methods (`@showLogin`)
  - Router calls (`Router::get(...)`)
- **Find References**
- **Peek Definition**

### 💡 Intelligent Hover
- Method signatures and documentation
- Route information with validation
- Processed docstrings (no asterisks)
- Fuzzy suggestions for typos

### 🔧 Diagnostics & Code Actions
- Real-time error detection
- Controller/method not found
- Security vulnerabilities
- Quick fixes and code actions

### ⚡ Commands (Ctrl+Shift+P)
- `Joss: Index Workspace`
- `Joss: Go to Route`
- `Joss: Make Controller`
- `Joss: Make Model`
- `Joss: Make CRUD`
- `Joss: Remove CRUD`
- `Joss: Make Migration`
- `Joss: Run Migrations`
- `Joss: Start Server`
- `Joss: New Project`
- `Joss: Run JosSecurity Check`
- `Joss: Open Definition Under Cursor`
- `Joss: Restart Language Server`

### 🛡️ Security Analysis
- 10+ security rules
- SQL injection detection
- Weak encryption detection
- Unsafe eval() usage
- Public route validation

## Installation

### From Source

```bash
cd vscode-joss
npm install
npm run compile
```

### Package Extension

```bash
npm run package
# Install joss-language-2.0.0.vsix in VS Code
```

## Configuration

```json
{
  "joss.indexOnOpen": true,
  "joss.maxFilesToIndex": 10000,
  "joss.enableJosSecurity": true,
  "joss.securitySeverity": "warning",
  "joss.controllerPaths": ["app/controllers"],
  "joss.modelPaths": ["app/models"]
}
```

## Usage

### Go-to-Definition

```joss
Router::get("/login", "AuthController@showLogin")
                      ^^^^^^^^^^^^^^ Ctrl+Click here
```

### Hover Information

Hover over any method or controller to see:
- Signature
- Location
- Documentation
- Validation status

### Security Check

Run `Joss: Run JosSecurity Check` to analyze your entire workspace for security issues.

## Development

```bash
# Install dependencies
npm install

# Compile TypeScript
npm run compile

# Watch mode
npm run watch

# Run tests
npm test

# Lint
npm run lint
```

## Architecture

```
vscode-joss/
├── src/
│   ├── extension.ts          # Client extension
│   └── server/
│       ├── server.ts          # Language server
│       ├── parser/
│       │   └── routeParser.ts # Route parsing
│       ├── indexer/
│       │   └── indexer.ts     # Symbol indexing
│       └── analyzer/
│           └── securityAnalyzer.ts # Security rules
├── syntaxes/                  # TextMate grammar
├── snippets/                  # Code snippets
└── themes/                    # Color themes
```

## License

MIT

## Version

3.0.1 - Added comprehensive commands, intelligent snippets (if/else/switch -> ternaries), and `remove:crud` support.
3.0.0 - Added support for Pipe Operator (`|>`) and JosSecurity v3.0.1.
2.0.0 - Complete LSP rewrite with advanced features

