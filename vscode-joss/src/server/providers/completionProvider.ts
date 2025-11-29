import { TextDocumentPositionParams, CompletionItem, CompletionItemKind } from 'vscode-languageserver/node';
import { connection, indexer } from '../server';

export function setupCompletionProvider() {
    connection.onCompletion(
        async (_textDocumentPosition: TextDocumentPositionParams): Promise<CompletionItem[]> => {
            const symbols = await indexer.getAllSymbols();

            return symbols.map(symbol => ({
                label: symbol.name,
                kind: symbol.kind === 'class' ? CompletionItemKind.Class : CompletionItemKind.Method,
                detail: symbol.signature,
                documentation: symbol.docstring
            }));
        }
    );
}
