import { Location, Range, SymbolKind } from 'vscode-languageserver/node';
import * as levenshtein from 'fast-levenshtein';

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
