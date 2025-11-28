const vscode = require('vscode');

function refreshDiagnostics(doc, diagnosticCollection) {
    const diagnostics = [];

    for (let lineIndex = 0; lineIndex < doc.lineCount; lineIndex++) {
        const lineOfText = doc.lineAt(lineIndex);
        const text = lineOfText.text.trim();

        // 1. Check Router::get/post arguments
        // Regex to find Router::get("path") without handler
        // Matches: Router::get("...") but not Router::get("...", "...")
        if (text.startsWith('Router::')) {
            const methodMatch = text.match(/Router::(get|post|put|delete|match)\s*\(([^)]+)\)/);
            if (methodMatch) {
                const args = methodMatch[2].split(',');
                if (methodMatch[1] === 'match') {
                    if (args.length < 3) {
                        diagnostics.push(createDiagnostic(doc, lineOfText, lineIndex, 'Router::match requires 3 arguments: (methods, path, handler)'));
                    }
                } else {
                    if (args.length < 2) {
                        diagnostics.push(createDiagnostic(doc, lineOfText, lineIndex, `Router::${methodMatch[1]} requires at least 2 arguments: (path, handler)`));
                    }
                }
            }
        }

        // 2. Check Class Definition
        // Should be: class Name { or class Name extends Parent {
        if (text.startsWith('class ')) {
            if (!text.includes('{')) {
                diagnostics.push(createDiagnostic(doc, lineOfText, lineIndex, 'Class definition missing opening brace "{"'));
            }
        }

        // 3. Check Function Definition
        // Should be: func name() { or Init name() {
        if (text.startsWith('func ') || text.startsWith('Init ')) {
            if (!text.includes('(') || !text.includes(')')) {
                diagnostics.push(createDiagnostic(doc, lineOfText, lineIndex, 'Function definition missing parentheses "()"'));
            }
            if (!text.includes('{')) {
                diagnostics.push(createDiagnostic(doc, lineOfText, lineIndex, 'Function definition missing opening brace "{"'));
            }
        }
    }

    diagnosticCollection.set(doc.uri, diagnostics);
}

function createDiagnostic(doc, lineOfText, lineIndex, message) {
    const range = new vscode.Range(lineIndex, 0, lineIndex, lineOfText.text.length);
    const diagnostic = new vscode.Diagnostic(range, message, vscode.DiagnosticSeverity.Error);
    return diagnostic;
}

function subscribeToDocumentChanges(context, diagnosticCollection) {
    if (vscode.window.activeTextEditor) {
        refreshDiagnostics(vscode.window.activeTextEditor.document, diagnosticCollection);
    }
    context.subscriptions.push(
        vscode.window.onDidChangeActiveTextEditor(editor => {
            if (editor) {
                refreshDiagnostics(editor.document, diagnosticCollection);
            }
        })
    );
    context.subscriptions.push(
        vscode.workspace.onDidChangeTextDocument(e => {
            refreshDiagnostics(e.document, diagnosticCollection);
        })
    );
    context.subscriptions.push(
        vscode.workspace.onDidCloseTextDocument(doc => diagnosticCollection.delete(doc.uri))
    );
}

module.exports = {
    subscribeToDocumentChanges
};
