import * as path from 'path';
import * as vscode from 'vscode';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind
} from 'vscode-languageclient/node';
import * as cp from 'child_process';

import { registerCommands } from './commands';

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext) {
    console.log('JosSecurity Extension v3.0 is now active!');

    // Start Language Server
    client = startLanguageServer(context);

    // Register Commands
    registerCommands(context, client);

    // Status Bar
    const statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    statusBarItem.text = '$(database) Joss';
    statusBarItem.tooltip = 'JosSecurity Language Server';
    statusBarItem.show();
    context.subscriptions.push(statusBarItem);
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    return client.stop();
}

function startLanguageServer(context: vscode.ExtensionContext): LanguageClient {
    // Server module path
    const serverModule = context.asAbsolutePath(
        path.join('out', 'server', 'server.js')
    );

    // Server options
    const serverOptions: ServerOptions = {
        run: { module: serverModule, transport: TransportKind.ipc },
        debug: {
            module: serverModule,
            transport: TransportKind.ipc,
            options: { execArgv: ['--nolazy', '--inspect=6009'] }
        }
    };

    // Client options
    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: 'file', language: 'joss' }],
        synchronize: {
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.joss')
        }
    };

    // Create and start client
    const client = new LanguageClient(
        'jossLanguageServer',
        'JosSecurity Language Server',
        serverOptions,
        clientOptions
    );

    client.start();

    return client;
}
