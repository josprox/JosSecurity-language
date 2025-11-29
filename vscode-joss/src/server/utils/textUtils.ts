import { TextDocument } from 'vscode-languageserver-textdocument';
import { Position } from 'vscode-languageserver/node';

export function getWordAtPosition(document: TextDocument, position: Position): string {
    const text = document.getText();
    const offset = document.offsetAt(position);

    let start = offset;
    let end = offset;

    while (start > 0 && /[a-zA-Z0-9_]/.test(text[start - 1])) {
        start--;
    }

    while (end < text.length && /[a-zA-Z0-9_]/.test(text[end])) {
        end++;
    }

    return text.substring(start, end);
}
