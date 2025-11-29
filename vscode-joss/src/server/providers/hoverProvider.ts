import { HoverParams, Hover, MarkupKind } from 'vscode-languageserver/node';
import { connection, documents, indexer } from '../server';
import { getWordAtPosition } from '../utils/textUtils';

export function setupHoverProvider() {
    connection.onHover(async (params: HoverParams): Promise<Hover | null> => {
        const document = documents.get(params.textDocument.uri);
        if (!document) return null;

        const position = params.position;
        const word = getWordAtPosition(document, position);

        // Find symbol
        const symbol = await indexer.findSymbol(word);
        if (!symbol) return null;

        // Build hover content
        const docstring = symbol.docstring || '';
        const content = `**${symbol.name}**${symbol.signature || ''}\n\n` +
            `*${symbol.location.uri}:${symbol.location.range.start.line + 1}*\n\n` +
            docstring;

        return {
            contents: {
                kind: MarkupKind.Markdown,
                value: content
            }
        };
    });
}
