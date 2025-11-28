const vscode = require('vscode');
const completion = require('./src/completion');
const hover = require('./src/hover');
const diagnostics = require('./src/diagnostics');

function activate(context) {
    console.log('JosSecurity Extension is now active!');

    // 1. Autocomplete Provider
    context.subscriptions.push(completion.getCompletionItemProvider());

    // 2. Hover Provider
    context.subscriptions.push(hover.getHoverProvider());

    // 3. Diagnostics (Syntax Checking)
    const diagnosticCollection = vscode.languages.createDiagnosticCollection('joss');
    context.subscriptions.push(diagnosticCollection);
    diagnostics.subscribeToDocumentChanges(context, diagnosticCollection);
}

function deactivate() { }

module.exports = {
    activate,
    deactivate
};
