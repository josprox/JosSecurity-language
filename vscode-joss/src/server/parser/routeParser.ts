import { Range, Position } from 'vscode-languageserver/node';

export interface ParsedRoute {
    method: string;
    path: string;
    controller: string;
    methods: Array<{ name: string; range: Range }>;
    controllerRange: Range;
    fullLine: number;
}

export class RouteParser {
    private readonly ROUTE_PATTERN = /Router::(get|post|put|delete|match|api)\s*\(\s*["']([^"']+)["'](?:\s*,\s*["']([^"']+)["'])?\s*,\s*["']([^"'@]+)@([^"']+)["']\s*\)/g;

    parseDocument(text: string): ParsedRoute[] {
        const routes: ParsedRoute[] = [];
        const lines = text.split('\n');

        for (let lineNum = 0; lineNum < lines.length; lineNum++) {
            const line = lines[lineNum];
            const parsed = this.parseLine(line, lineNum);
            if (parsed) {
                routes.push(parsed);
            }
        }

        return routes;
    }

    private parseLine(line: string, lineNum: number): ParsedRoute | null {
        const match = this.ROUTE_PATTERN.exec(line);
        if (!match) return null;

        const [fullMatch, method, path, , controller, methodsStr] = match;
        const methods = methodsStr.split('@').map((m, i) => ({
            name: m,
            range: this.getRangeInLine(line, lineNum, m)
        }));

        return {
            method,
            path,
            controller,
            methods,
            controllerRange: this.getRangeInLine(line, lineNum, controller),
            fullLine: lineNum
        };
    }

    private getRangeInLine(line: string, lineNum: number, word: string): Range {
        const start = line.indexOf(word);
        return {
            start: { line: lineNum, character: start },
            end: { line: lineNum, character: start + word.length }
        };
    }
}
