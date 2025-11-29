import { connection, indexer } from '../server';

export function setupCustomRequests() {
    connection.onRequest('joss/getRoutes', async () => {
        return await indexer.getAllRoutes();
    });

    connection.onRequest('joss/resolveRoute', async (route: any) => {
        const controller = await indexer.findController(route.controller);
        return controller;
    });

    connection.onRequest('joss/createController', async (params: { name: string }) => {
        // TODO: Implement controller creation
        return { success: true, path: `app/controllers/${params.name}.joss` };
    });

    connection.onRequest('joss/securityCheck', async () => {
        // TODO: Implement full workspace security check
        return { issues: [] };
    });
}
