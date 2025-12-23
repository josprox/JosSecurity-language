import { Location, Range, SymbolKind, Position } from 'vscode-languageserver/node';
import * as levenshtein from 'fast-levenshtein';
import * as fs from 'fs';
import * as path from 'path';
import { URI } from 'vscode-uri';

export interface Symbol {
    name: string;
    kind: string;
    location: Location;
    signature?: string;
    docstring?: string;
}

export class Indexer {
    private symbols: Map<string, Symbol[]> = new Map();
    private routes: any[] = [];
    private workspaceRoot: string = '';

    setWorkspaceRoot(root: string) {
        this.workspaceRoot = root;
    }

    async indexWorkspace(): Promise<void> {
        if (!this.workspaceRoot) return;

        this.symbols.clear();
        this.routes = [];

        await this.scanDirectory(this.workspaceRoot);
    }

    private async scanDirectory(dir: string): Promise<void> {
        try {
            const entries = fs.readdirSync(dir, { withFileTypes: true });

            for (const entry of entries) {
                const fullPath = path.join(dir, entry.name);

                if (entry.isDirectory()) {
                    // Skip node_modules, .git, etc.
                    if (!['node_modules', '.git', 'out', 'dist'].includes(entry.name)) {
                        await this.scanDirectory(fullPath);
                    }
                } else if (entry.isFile() && entry.name.endsWith('.joss')) {
                    await this.indexFile(fullPath);
                }
            }
        } catch (error) {
            console.error(`Error scanning directory ${dir}:`, error);
        }
    }

    private async indexFile(filePath: string): Promise<void> {
        try {
            const content = fs.readFileSync(filePath, 'utf-8');
            const uri = URI.file(filePath).toString();

            // Extract classes
            const classRegex = /class\s+(\w+)/g;
            let match;
            while ((match = classRegex.exec(content)) !== null) {
                const className = match[1];
                const line = content.substring(0, match.index).split('\n').length - 1;

                this.addSymbol({
                    name: className,
                    kind: 'class',
                    location: {
                        uri,
                        range: {
                            start: { line, character: match.index },
                            end: { line, character: match.index + match[0].length }
                        }
                    }
                });

                // Extract methods within this class
                const classStart = match.index;
                const classEnd = this.findClassEnd(content, classStart);
                const classContent = content.substring(classStart, classEnd);

                const methodRegex = /(?:function|func)\s+(\w+)\s*\(/g;
                let methodMatch;
                while ((methodMatch = methodRegex.exec(classContent)) !== null) {
                    const methodName = methodMatch[1];
                    const methodLine = content.substring(0, classStart + methodMatch.index).split('\n').length - 1;

                    this.addSymbol({
                        name: `${className}.${methodName}`,
                        kind: 'method',
                        location: {
                            uri,
                            range: {
                                start: { line: methodLine, character: methodMatch.index },
                                end: { line: methodLine, character: methodMatch.index + methodMatch[0].length }
                            }
                        }
                    });
                }
            }

            // Extract standalone functions
            const functionRegex = /^(?:function|func)\s+(\w+)\s*\(/gm;
            while ((match = functionRegex.exec(content)) !== null) {
                const functionName = match[1];
                const line = content.substring(0, match.index).split('\n').length - 1;

                this.addSymbol({
                    name: functionName,
                    kind: 'function',
                    location: {
                        uri,
                        range: {
                            start: { line, character: match.index },
                            end: { line, character: match.index + match[0].length }
                        }
                    }
                });
            }

            // Extract routes
            const routeRegex = /Router(?:::|\.)(get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']\s*,\s*["'](\w+)@(\w+)["']/g;
            while ((match = routeRegex.exec(content)) !== null) {
                const [_, method, routePath, controller, action] = match;
                const line = content.substring(0, match.index).split('\n').length - 1;

                this.addRoute({
                    method: method.toUpperCase(),
                    path: routePath,
                    controller,
                    action,
                    location: {
                        uri,
                        range: {
                            start: { line, character: match.index },
                            end: { line, character: match.index + match[0].length }
                        }
                    }
                });
            }
        } catch (error) {
            console.error(`Error indexing file ${filePath}:`, error);
        }
    }

    private findClassEnd(content: string, classStart: number): number {
        let braceCount = 0;
        let inClass = false;

        for (let i = classStart; i < content.length; i++) {
            if (content[i] === '{') {
                braceCount++;
                inClass = true;
            } else if (content[i] === '}') {
                braceCount--;
                if (inClass && braceCount === 0) {
                    return i + 1;
                }
            }
        }

        return content.length;
    }

    async findSymbol(name: string): Promise<Symbol | null> {
        const symbols = this.symbols.get(name);
        return symbols?.[0] || null;
    }

    async findController(name: string): Promise<Location | null> {
        const symbol = await this.findSymbol(name);
        return symbol?.location || null;
    }

    async findMethod(controller: string, method: string): Promise<Symbol | null> {
        const key = `${controller}.${method}`;
        return await this.findSymbol(key);
    }

    async fuzzyFindMethod(controller: string, method: string): Promise<string[]> {
        const allMethods = Array.from(this.symbols.keys())
            .filter(k => k.startsWith(controller + '.'))
            .map(k => k.split('.')[1]);

        return allMethods
            .map(m => ({ name: m, distance: levenshtein.get(method, m) }))
            .filter(m => m.distance <= 3)
            .sort((a, b) => a.distance - b.distance)
            .map(m => m.name);
    }

    async getAllSymbols(): Promise<Symbol[]> {
        const all: Symbol[] = [];
        for (const symbols of this.symbols.values()) {
            all.push(...symbols);
        }
        return all;
    }

    async getAllRoutes(): Promise<any[]> {
        return this.routes;
    }

    addSymbol(symbol: Symbol) {
        if (!this.symbols.has(symbol.name)) {
            this.symbols.set(symbol.name, []);
        }
        this.symbols.get(symbol.name)!.push(symbol);
    }

    addRoute(route: any) {
        this.routes.push(route);
    }
}
