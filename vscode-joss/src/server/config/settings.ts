export interface JossSettings {
    indexOnOpen: boolean;
    maxFilesToIndex: number;
    enableJosSecurity: boolean;
    securitySeverity: string;
    controllerPaths: string[];
    modelPaths: string[];
}

export function getDefaultSettings(): JossSettings {
    return {
        indexOnOpen: true,
        maxFilesToIndex: 10000,
        enableJosSecurity: true,
        securitySeverity: 'warning',
        controllerPaths: ['app/controllers', 'app/Controllers'],
        modelPaths: ['app/models', 'app/Models']
    };
}
