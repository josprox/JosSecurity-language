# Changelog

## [2.0.0] - 2025-11-28

### Added
- **Language Server Protocol (LSP)** implementation
- **Go-to-Definition** for controllers and methods
- **Intelligent Hover** with processed docstrings
- **Real-time Diagnostics** with code actions
- **Security Analysis** with 10+ rules
- **Workspace Indexing** with incremental updates
- **6 New Commands** via Ctrl+Shift+P
- **Fuzzy Search** for method suggestions
- **Route Navigation** with Quick Pick
- **Controller Stub Generation**

### Changed
- Complete rewrite from JavaScript to TypeScript
- Migrated from basic providers to full LSP
- Improved performance with caching
- Better error messages with suggestions

### Technical
- TypeScript 5.0
- vscode-languageclient 9.0
- vscode-languageserver 9.0
- LevelDB for caching
- Fast-Levenshtein for fuzzy matching

## [1.3.0] - Previous

### Added
- Router::match support
- Auth module snippets
- Middleware syntax highlighting

## [1.0.0] - Initial

### Added
- Basic syntax highlighting
- Code snippets
- JosSecurity Dark theme
