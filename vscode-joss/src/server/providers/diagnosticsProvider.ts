import { Diagnostic, DiagnosticSeverity } from 'vscode-languageserver/node';
import { TextDocument } from 'vscode-languageserver-textdocument';
import { connection, documents, routeParser, indexer, securityAnalyzer } from '../server';
import { getDocumentSettings } from '../server';

export function setupDiagnostics() {
    documents.onDidChangeContent(change => {
        validateTextDocument(change.document);
    });
}

async function validateTextDocument(textDocument: TextDocument): Promise<void> {
    const settings = await getDocumentSettings(textDocument.uri);
    const text = textDocument.getText();
    const diagnostics: Diagnostic[] = [];

    // Parse routes and validate
    const routes = routeParser.parseDocument(text);

    for (const route of routes) {
        // Validate controller exists
        const controller = await indexer.findController(route.controller);
        if (!controller) {
            diagnostics.push({
                severity: DiagnosticSeverity.Error,
                range: route.controllerRange,
                message: `Controller '${route.controller}' not found`,
                source: 'joss'
            });
        } else {
            // Validate methods exist
            for (const method of route.methods) {
                const methodSymbol = await indexer.findMethod(route.controller, method.name);
                if (!methodSymbol) {
                    const suggestions = await indexer.fuzzyFindMethod(route.controller, method.name);
                    const suggestionText = suggestions.length > 0
                        ? `\nDid you mean '${suggestions[0]}'?`
                        : '';

                    diagnostics.push({
                        severity: DiagnosticSeverity.Error,
                        range: method.range,
                        message: `Method '${method.name}' not found in ${route.controller}${suggestionText}`,
                        source: 'joss'
                    });
                }
            }
        }
    }

    // Security analysis
    if (settings.enableJosSecurity) {
        const securityIssues = await securityAnalyzer.analyze(text);
        for (const issue of securityIssues) {
            diagnostics.push({
                severity: issue.severity === 'error'
                    ? DiagnosticSeverity.Error
                    : DiagnosticSeverity.Warning,
                range: issue.range,
                message: issue.message,
                source: 'joss-security'
            });
        }
    }

    connection.sendDiagnostics({ uri: textDocument.uri, diagnostics });
}
