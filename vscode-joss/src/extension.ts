import * as path from 'path';
import * as vscode from 'vscode';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind
} from 'vscode-languageclient/node';
import * as cp from 'child_process';

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext) {
    console.log('JosSecurity Extension v2.0 is now active!');

    // Start Language Server
    client = startLanguageServer(context);

    // Register Commands
    registerCommands(context);

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

function registerCommands(context: vscode.ExtensionContext) {
    // Index Workspace
    context.subscriptions.push(
        vscode.commands.registerCommand('joss.indexWorkspace', async () => {
            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: 'Indexing JosSecurity workspace...',
                cancellable: false
            }, async (progress) => {
                progress.report({ increment: 0 });

                // Send command to server
                await client.sendRequest('workspace/executeCommand', {
                    command: 'joss.indexWorkspace'
                });

                progress.report({ increment: 100 });
                vscode.window.showInformationMessage('Workspace indexed successfully!');
            });
        })
    );

    // Go to Route
    context.subscriptions.push(
        vscode.commands.registerCommand('joss.goToRoute', async () => {
            const routes = await client.sendRequest<any[]>('joss/getRoutes');

            if (!routes || routes.length === 0) {
                vscode.window.showInformationMessage('No routes found');
                return;
            }

            const items = routes.map(r => ({
                label: `${r.method} ${r.path}`,
                description: `${r.controller}@${r.methods.join('@')}`,
                route: r
            }));

            const selected = await vscode.window.showQuickPick(items, {
                placeHolder: 'Select a route to navigate to'
            });

            if (selected) {
                const location = await client.sendRequest<any>('joss/resolveRoute', selected.route);
                if (location) {
                    const uri = vscode.Uri.parse(location.uri);
                    const position = new vscode.Position(location.range.start.line, location.range.start.character);
                    await vscode.window.showTextDocument(uri, { selection: new vscode.Range(position, position) });
                }
            }
        })
    );

    // Create Controller
    context.subscriptions.push(
        vscode.commands.registerCommand('joss.createController', async () => {
            const name = await vscode.window.showInputBox({
                prompt: 'Controller name (e.g., UserController)',
                placeHolder: 'UserController',
                validateInput: (value) => {
                    if (!value) return 'Controller name is required';
                    if (!value.endsWith('Controller')) return 'Controller name should end with "Controller"';
                    return null;
                }
            });

            if (name) {
                const result = await client.sendRequest<any>('joss/createController', { name });
                if (result.success) {
                    vscode.window.showInformationMessage(`Controller ${name} created successfully!`);
                    const uri = vscode.Uri.file(result.path);
                    await vscode.window.showTextDocument(uri);
                } else {
                    vscode.window.showErrorMessage(`Failed to create controller: ${result.error}`);
                }
            }
        })
    );

    // Run Security Check
    context.subscriptions.push(
        vscode.commands.registerCommand('joss.runSecurityCheck', async () => {
            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: 'Running JosSecurity analysis...',
                cancellable: false
            }, async (progress) => {
                progress.report({ increment: 0 });

                const results = await client.sendRequest<any>('joss/securityCheck');

                progress.report({ increment: 100 });

                if (results.issues.length === 0) {
                    vscode.window.showInformationMessage('✓ No security issues found!');
                } else {
                    vscode.window.showWarningMessage(`Found ${results.issues.length} security issue(s)`);
                }
            });
        })
    );

    // Open Definition
    context.subscriptions.push(
        vscode.commands.registerCommand('joss.openDefinition', async () => {
            await vscode.commands.executeCommand('editor.action.revealDefinition');
        })
    );

    // Restart Server
    context.subscriptions.push(
        vscode.commands.registerCommand('joss.restartServer', async () => {
            await client.stop();
            client.start();
            vscode.window.showInformationMessage('Language server restarted');
        })
    );

    // Change DB Prefix
    context.subscriptions.push(
        vscode.commands.registerCommand('joss.changeDbPrefix', async () => {
            const prefix = await vscode.window.showInputBox({
                prompt: 'Enter new database prefix (e.g., app_)',
                placeHolder: 'app_',
                validateInput: (value) => {
                    if (!value) return 'Prefix is required';
                    return null;
                }
            });

            if (prefix) {
                const workspaceFolders = vscode.workspace.workspaceFolders;
                if (!workspaceFolders) {
                    vscode.window.showErrorMessage('No workspace open');
                    return;
                }
                const rootPath = workspaceFolders[0].uri.fsPath;

                vscode.window.withProgress({
                    location: vscode.ProgressLocation.Notification,
                    title: `Changing DB prefix to '${prefix}'...`,
                    cancellable: false
                }, async (progress) => {
                    return new Promise<void>((resolve, reject) => {
                        cp.exec(`joss change db prefix ${prefix}`, { cwd: rootPath }, (err, stdout, stderr) => {
                            if (err) {
                                vscode.window.showErrorMessage(`Failed to change prefix: ${stderr || err.message}`);
                                resolve();
                            } else {
                                vscode.window.showInformationMessage(`Database prefix changed to '${prefix}' successfully!`);
                                resolve();
                            }
                        });
                    });
                });
            }
        })
    );
}
