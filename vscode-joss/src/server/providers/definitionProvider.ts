import { DefinitionParams, Definition, Location } from 'vscode-languageserver/node';
import { connection, documents, indexer } from '../server';
import { referenceAtPosition } from '../utils/callContext';
import * as path from 'path';
import * as fs from 'fs';
import { URI } from 'vscode-uri';

export function setupDefinitionProvider() {
    connection.onDefinition(async (params: DefinitionParams): Promise<Definition | null> => {
        const document = documents.get(params.textDocument.uri);
        if (!document) return null;

        const position = params.position;
        const line = document.getText({
            start: { line: position.line, character: 0 },
            end: { line: position.line, character: 1000 }
        });

        // 1. Router definitions: Router::get(... "Controller@method")
        const routerMatch = line.match(/Router::(get|post|put|delete|patch|options|api|match)\s*\([\s\S]*?["']([A-Za-z0-9_]+)@([A-Za-z0-9_]+)["']/);
        if (routerMatch) {
            const controller = routerMatch[2];
            const action = routerMatch[3];

            const controllerStart = line.indexOf(controller);
            const controllerEnd = controllerStart + controller.length;
            if (position.character >= controllerStart && position.character <= controllerEnd) {
                const symbol = await indexer.findSymbol(controller);
                if (symbol) return symbol.location;
            }

            const methodStart = line.indexOf(action, controllerStart);
            const methodEnd = methodStart + action.length;
            if (position.character >= methodStart && position.character <= methodEnd) {
                const symbol = await indexer.findMethod(controller, action);
                if (symbol) return symbol.location;
            }
        }

        // 2. View definitions: View::render("users.profile") or View::render("users/profile")
        const viewMatch = line.match(/View::render\s*\(\s*["']([^"']+)["']/);
        if (viewMatch) {
            const viewRaw = viewMatch[1];
            const viewStart = line.indexOf(viewRaw);
            const viewEnd = viewStart + viewRaw.length;
            if (position.character >= viewStart && position.character <= viewEnd) {
                const location = resolveViewLocation(viewRaw, indexer.getWorkspaceRoot());
                if (location) return location;
            }
        }

        // 3. View template directives: @extends("layouts.app") or @include("partials.header")
        const templateMatch = line.match(/@(extends|include|section)\s*\(\s*["']([^"']+)["']/);
        if (templateMatch) {
            const viewRaw = templateMatch[2];
            const viewStart = line.indexOf(viewRaw);
            const viewEnd = viewStart + viewRaw.length;
            if (position.character >= viewStart && position.character <= viewEnd) {
                const location = resolveViewLocation(viewRaw, indexer.getWorkspaceRoot());
                if (location) return location;
            }
        }

        // Fallback: try to find symbol at cursor position
        const word = referenceAtPosition(document, position);

        // Try direct symbol lookup
        let symbol = await indexer.findSymbol(word);
        if (symbol) return symbol.location;

        // Try as Controller.method
        if (word.includes('::') || word.includes('->')) {
            const [controller, method] = word.split(/::|->/);
            symbol = await indexer.findMethod(controller, method);
            if (symbol) return symbol.location;
        }

        const sameName = await indexer.findSymbolsBySimpleName(word.replace(/^\$/, ''));
        if (sameName.length) return sameName[0].location;

        return null;
    });
}

function resolveViewLocation(viewName: string, workspaceRoot: string): Location | null {
    if (!workspaceRoot) return null;
    const normalized = viewName.replace(/\./g, '/').replace(/^\/+/, '');
    const candidatePaths = [
        path.join(workspaceRoot, 'app', 'views', `${normalized}.joss.html`),
        path.join(workspaceRoot, 'app', 'views', `${normalized}.html`),
        path.join(workspaceRoot, 'views', `${normalized}.joss.html`),
        path.join(workspaceRoot, 'views', `${normalized}.html`),
        path.join(workspaceRoot, 'resources', 'views', `${normalized}.joss.html`)
    ];

    for (const candidate of candidatePaths) {
        if (fs.existsSync(candidate)) {
            return {
                uri: URI.file(candidate).toString(),
                range: {
                    start: { line: 0, character: 0 },
                    end: { line: 0, character: 0 }
                }
            };
        }
    }
    return null;
}
