# Contributing to JosSecurity VS Code Extension

## Development Setup

1. **Clone and Install**
```bash
git clone https://github.com/jossecurity/joss
cd vscode-joss
npm install
```

2. **Compile TypeScript**
```bash
npm run compile
# or watch mode
npm run watch
```

3. **Debug Extension**
- Open `vscode-joss` in VS Code
- Press F5 to launch Extension Development Host
- Test features in the new window

## Project Structure

```
vscode-joss/
├── src/
│   ├── extension.ts              # Client extension entry point
│   └── server/
│       ├── server.ts              # Language server
│       ├── parser/
│       │   └── routeParser.ts     # Parse Router calls
│       ├── indexer/
│       │   └── indexer.ts         # Symbol indexing
│       └── analyzer/
│           └── securityAnalyzer.ts # Security rules
├── syntaxes/                      # TextMate grammar
├── snippets/                      # Code snippets
├── themes/                        # Color themes
└── test-files/                    # Example files for testing
```

## Adding Features

### New Security Rule

Edit `src/server/analyzer/securityAnalyzer.ts`:

```typescript
{
  id: 'my-rule',
  pattern: /dangerous_pattern/g,
  severity: 'error',
  message: 'Description of issue'
}
```

### New Command

1. Add to `package.json` contributes.commands
2. Register in `src/extension.ts`
3. Implement handler

### New LSP Feature

Implement in `src/server/server.ts`:
- `connection.onHover()`
- `connection.onDefinition()`
- `connection.onCompletion()`
- etc.

## Testing

### Manual Testing

1. Open test-files/example.joss
2. Test features:
   - Hover over methods
   - Ctrl+Click on controllers
   - Check diagnostics
   - Run commands (Ctrl+Shift+P)

### Automated Tests

```bash
npm test
```

## Code Style

- Use TypeScript strict mode
- Follow ESLint rules
- Add JSDoc comments
- Keep functions small and focused

## Pull Requests

1. Create feature branch
2. Make changes
3. Test thoroughly
4. Submit PR with description

## Questions?

Open an issue on GitHub
