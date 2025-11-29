import { DefinitionParams, Definition } from 'vscode-languageserver/node';
import { connection, documents, indexer } from '../server';
import { getWordAtPosition } from '../utils/textUtils';

export function setupDefinitionProvider() {
    connection.onDefinition(async (params: DefinitionParams): Promise<Definition | null> => {
        const document = documents.get(params.textDocument.uri);
        if (!document) return null;

        const position = params.position;
        const word = getWordAtPosition(document, position);

        // Find symbol
        const symbol = await indexer.findSymbol(word);
        if (!symbol) return null;

        return symbol.location;
    });
}
