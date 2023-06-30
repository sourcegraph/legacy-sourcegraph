import vscode, { Uri } from 'vscode'

import { LightTextDocument, CompletionsTextEditor } from '@sourcegraph/cody-shared/src/autocomplete'

export const textEditor: CompletionsTextEditor = {
    getWorkspaceRootPath(): string | null {
        const uri = vscode.window.activeTextEditor?.document?.uri
        if (uri) {
            const wsFolder = vscode.workspace.getWorkspaceFolder(uri)
            if (wsFolder) {
                return wsFolder.uri.fsPath
            }
        }
        return vscode.workspace.workspaceFolders?.[0]?.uri?.fsPath ?? null
    },

    getOpenDocuments(): LightTextDocument[] {
        return vscode.workspace.textDocuments.map(doc => ({
            uri: doc.uri.toString(),
            languageId: doc.languageId,
        }))
    },

    getCurrentDocument(): LightTextDocument | null {
        const curr = vscode.window.activeTextEditor?.document
        if (!curr) {
            return null
        }

        return {
            uri: curr.uri.toString(),
            languageId: curr.languageId,
        }
    },

    async getDocumentTextTruncated(uri: string): Promise<string | null> {
        const document = await vscode.workspace.openTextDocument(Uri.parse(uri))
        const endLine = Math.min(document.lineCount, 10_000)
        const range = new vscode.Range(0, 0, endLine, 0)
        return document.getText(range)
    },

    async getDocumentRelativePath(uri: string): Promise<string | null> {
        return vscode.workspace.asRelativePath((await vscode.workspace.openTextDocument(Uri.parse(uri))).uri.fsPath)
    },

    getTabSize() {
        return vscode.window.activeTextEditor
            ? // tabSize is always resolved to a number when accessing the property
              (vscode.window.activeTextEditor.options.tabSize as number)
            : 2
    },
}
