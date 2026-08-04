import { Hover, HoverParams, MarkupKind } from 'vscode-languageserver/node';
import { connection, documents, indexer } from '../server';
import { findNativeCallable, nativeClasses, nativeSignature } from '../nativeCatalog';
import { referenceAtPosition } from '../utils/callContext';

export function setupHoverProvider() {
    connection.onHover(async (params: HoverParams): Promise<Hover | null> => {
        const document = documents.get(params.textDocument.uri);
        if (!document) return null;
        const reference = referenceAtPosition(document, params.position);

        if (['use', 'import', '@import'].includes(reference.toLowerCase())) {
            return markdown(`⚠️ **Sintaxis Obsoleta en Joss**\n\nLas instrucciones \`use\` e \`import\` fueron eliminadas del lenguaje. Los plugins y paquetes declarados en \`joss.yaml\` se cargan e indexan automáticamente en el espacio de nombres global.`);
        }

        const native = findNativeCallable(reference);
        if (native) {
            return markdown(`\`\`\`joss\n${nativeSignature(native)}\n\`\`\`\n\n${native.documentation}${parameterDocs(native.parameters)}`);
        }
        if (nativeClasses.includes(reference)) {
            return markdown(`**${reference}**\n\nClase nativa de Joss. Escribe \`${reference}::\` para ver sus métodos.`);
        }

        let symbol = await indexer.findSymbol(reference);
        if (!symbol) {
            symbol = (await indexer.findSymbolsBySimpleName(reference.replace(/^\$/, '')))[0] || null;
        }
        if (!symbol) return null;
        const location = symbol.location.uri.replace('file:///', '');
        const signature = symbol.signature || symbol.qualifiedName;

        let originBadge = '';
        if (symbol.origin === 'plugin' || symbol.packageName) {
            const pkgName = symbol.packageName || 'plugin';
            const verStr = symbol.packageVersion ? ` (v${symbol.packageVersion})` : '';
            originBadge = `\n\n📦 **Plugin Originario**: \`${pkgName}\`${verStr}\n*Origen: Paquete autocontenido \`${pkgName}.jp\`*`;
        }

        return markdown(`\`\`\`joss\n${signature}\n\`\`\`${originBadge}\n\n${symbol.docstring || `Símbolo ${symbol.kind} del proyecto.`}${parameterDocs(symbol.parameters)}\n\n_${location}:${symbol.location.range.start.line + 1}_`);
    });
}

function parameterDocs(parameters: Array<{ label: string; documentation?: string }>): string {
    if (!parameters.length) return '';
    return `\n\n**Parámetros**\n${parameters.map(parameter => `- \`${parameter.label}\`${parameter.documentation ? ` — ${parameter.documentation}` : ''}`).join('\n')}`;
}

function markdown(value: string): Hover {
    return { contents: { kind: MarkupKind.Markdown, value } };
}
