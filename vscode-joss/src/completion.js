const vscode = require('vscode');

function getCompletionItemProvider() {
    return vscode.languages.registerCompletionItemProvider('joss', {
        provideCompletionItems(document, position, token, context) {
            const linePrefix = document.lineAt(position).text.substr(0, position.character);

            // --- Router ---
            if (linePrefix.endsWith('Router::')) {
                return [
                    createMethod('get', 'get(path, handler)', 'Defines a GET route.'),
                    createMethod('post', 'post(path, handler)', 'Defines a POST route.'),
                    createMethod('put', 'put(path, handler)', 'Defines a PUT route.'),
                    createMethod('delete', 'delete(path, handler)', 'Defines a DELETE route.'),
                    createMethod('match', 'match(methods, path, handler)', 'Defines a route for multiple methods (e.g. "GET|POST").'),
                    createMethod('middleware', 'middleware(name)', 'Starts a middleware group (e.g. "auth", "guest").'),
                    createMethod('end', 'end()', 'Ends the current middleware group.'),
                    createMethod('api', 'api(path, handler)', 'Defines an API route (GET & POST).')
                ];
            }

            // --- Auth ---
            if (linePrefix.endsWith('Auth::')) {
                return [
                    createMethod('check', 'check()', 'Returns true if user is authenticated.'),
                    createMethod('guest', 'guest()', 'Returns true if user is NOT authenticated.'),
                    createMethod('user', 'user()', 'Returns the current user name.'),
                    createMethod('id', 'id()', 'Returns the current user ID.'),
                    createMethod('hasRole', 'hasRole(role)', 'Checks if user has a specific role.'),
                    createMethod('attempt', 'attempt(email, password)', 'Attempts login. Returns true on success.'),
                    createMethod('create', 'create([email, password, name])', 'Creates a new user.'),
                    createMethod('logout', 'logout()', 'Logs out the current user.')
                ];
            }

            // --- Request ---
            if (linePrefix.endsWith('Request::')) {
                return [
                    createMethod('input', 'input(key, default)', 'Retrieves a value from request data.'),
                    createMethod('all', 'all()', 'Returns all request data.'),
                    createMethod('file', 'file(key)', 'Retrieves an uploaded file.')
                ];
            }

            // --- Response ---
            if (linePrefix.endsWith('Response::')) {
                return [
                    createMethod('json', 'json(data)', 'Returns a JSON response.'),
                    createMethod('redirect', 'redirect(url)', 'Redirects to a URL.'),
                    createMethod('back', 'back()', 'Redirects back to the previous page.')
                ];
            }

            // --- View ---
            if (linePrefix.endsWith('View::')) {
                return [
                    createMethod('render', 'render(viewName, data)', 'Renders an HTML view.')
                ];
            }

            // --- System ---
            if (linePrefix.endsWith('System.')) {
                return [
                    createMethod('Run', 'Run(command, args)', 'Executes a system command.'),
                    createMethod('env', 'env(key, default)', 'Retrieves an environment variable securely.')
                ];
            }

            // --- Global Functions & Keywords ---
            if (!linePrefix.includes('.')) {
                return [
                    createFunction('print', 'print(msg)', 'Prints to stdout.'),
                    createFunction('echo', 'echo(msg)', 'Alias for print.'),
                    createFunction('dd', 'dd(var)', 'Dump and Die.'),
                    createKeyword('class', 'Defines a new class.'),
                    createKeyword('func', 'Defines a new function.'),
                    createKeyword('Init', 'Defines an initializer/constructor.'),
                    createKeyword('var', 'Defines a variable.'),
                    createKeyword('const', 'Defines a constant.'),
                    createKeyword('return', 'Returns a value.'),
                    createKeyword('if', 'Conditional statement.'),
                    createKeyword('else', 'Else statement.'),
                    createKeyword('for', 'Loop statement.'),
                    createKeyword('foreach', 'Iterates over arrays/maps.'),
                    createKeyword('async', 'Starts an asynchronous operation.'),
                    createKeyword('await', 'Waits for an asynchronous operation.'),
                ];
            }

            return [];
        }
    }, '.', ':');
}

function createMethod(label, detail, doc) {
    const item = new vscode.CompletionItem(label, vscode.CompletionItemKind.Method);
    item.detail = detail;
    item.documentation = new vscode.MarkdownString(doc);
    return item;
}

function createFunction(label, detail, doc) {
    const item = new vscode.CompletionItem(label, vscode.CompletionItemKind.Function);
    item.detail = detail;
    item.documentation = new vscode.MarkdownString(doc);
    return item;
}

function createKeyword(label, doc) {
    const item = new vscode.CompletionItem(label, vscode.CompletionItemKind.Keyword);
    item.documentation = new vscode.MarkdownString(doc);
    return item;
}

module.exports = {
    getCompletionItemProvider
};
