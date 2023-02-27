import * as vscode from 'vscode'

import { ChatViewProvider } from './chat/view'
import { WSChatClient } from './chat/ws'
import { WSCompletionsClient, fetchAndShowCompletions } from './completions'
import { Configuration, ConfigurationUseContext, getConfiguration } from './configuration'
import { CompletionsDocumentProvider } from './docprovider'
import { EmbeddingsClient } from './embeddings-client'
import { History } from './history'

const CODY_ACCESS_TOKEN_SECRET = 'cody.access-token'

export async function activate(context: vscode.ExtensionContext): Promise<void> {
    console.log('Cody extension activated')

    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.cody.toggleEnabled', async () => {
            const config = vscode.workspace.getConfiguration()
            await config.update(
                'sourcegraph.cody.enable',
                !config.get('sourcegraph.cody.enable'),
                vscode.ConfigurationTarget.Global
            )
        }),
        vscode.commands.registerCommand('cody.set-access-token', async () => {
            const tokenInput = await vscode.window.showInputBox()
            if (tokenInput === undefined || tokenInput === '') {
                return
            }
            await context.secrets.store(CODY_ACCESS_TOKEN_SECRET, tokenInput)
        }),
        vscode.commands.registerCommand('cody.delete-access-token', async () =>
            context.secrets.delete(CODY_ACCESS_TOKEN_SECRET)
        )
    )

    let disposable: vscode.Disposable | undefined
    context.subscriptions.push({
        dispose: () => disposable?.dispose(),
    })
    const doConfigure = async (): Promise<void> => {
        disposable?.dispose()
        const config = getConfiguration(vscode.workspace.getConfiguration())
        const accessToken = (await context.secrets.get(CODY_ACCESS_TOKEN_SECRET)) ?? null
        disposable = configure(context, config, accessToken)
    }

    // Watch all relevant configuration and secrets for changes.
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration(async event => {
            if (event.affectsConfiguration('cody') || event.affectsConfiguration('sourcegraph')) {
                await doConfigure()
            }
        })
    )
    context.subscriptions.push(
        context.secrets.onDidChange(async event => {
            if (event.key === CODY_ACCESS_TOKEN_SECRET) {
                await doConfigure()
            }
        })
    )

    await doConfigure()
}

function configure(
    context: Pick<vscode.ExtensionContext, 'extensionPath' | 'secrets'>,
    config: Configuration,
    accessToken: string | null
): vscode.Disposable {
    if (!config.enable || !accessToken) {
        if (config.enable) {
            const SET_ACCESS_TOKEN_ITEM = 'Set Access Token' as const
            vscode.window
                .showWarningMessage('Cody requires an access token.', SET_ACCESS_TOKEN_ITEM)
                .then(async item => {
                    if (item === SET_ACCESS_TOKEN_ITEM) {
                        await vscode.commands.executeCommand('cody.set-access-token')
                    }
                }, undefined)
        }
        setContextActivated(false)
        return NOOP_DISPOSABLE
    }

    // TODO(sqs): remove this check after devs have migrated (probably by 2023-03-01)
    if (!config.serverEndpoint.startsWith('http') || !config.embeddingsEndpoint.startsWith('http')) {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        vscode.window.showWarningMessage(
            'The cody.{server,embeddings}Endpoint settings now expect URLs (eg http://localhost:9300 or https://cody.sgdev.org).'
        )
        setContextActivated(false)
        return NOOP_DISPOSABLE
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const subscriptions: { dispose(): any }[] = []

    const documentProvider = new CompletionsDocumentProvider()
    const history = new History()
    subscriptions.push(history)

    const wsCompletionsClient = WSCompletionsClient.new(`${config.serverEndpoint}/completions`, accessToken)
    const wsChatClient = WSChatClient.new(`${config.serverEndpoint}/chat`, accessToken)
    const embeddingsClient = config.codebase
        ? new EmbeddingsClient(config.embeddingsEndpoint, accessToken, config.codebase)
        : null

    let useContext: ConfigurationUseContext = config.useContext
    if (!embeddingsClient && config.useContext === 'embeddings') {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        vscode.window.showInformationMessage(
            'Embeddings were not available (is `cody.codebase` set?), falling back to keyword context'
        )
        useContext = 'keyword'
    }

    const chatProvider = new ChatViewProvider(
        context.extensionPath,
        config.serverEndpoint,
        accessToken,
        wsChatClient,
        embeddingsClient,
        useContext,
        config.debug
    )

    const executeRecipe = async (recipe: string): Promise<void> => {
        await vscode.commands.executeCommand('cody.chat.focus')
        await chatProvider.executeRecipe(recipe)
    }

    subscriptions.push(
        vscode.workspace.registerTextDocumentContentProvider('codegen', documentProvider),
        vscode.languages.registerHoverProvider({ scheme: 'codegen' }, documentProvider),

        vscode.commands.registerCommand('cody.recipe.explain-code', async () => executeRecipe('explainCode')),
        vscode.commands.registerCommand('cody.recipe.explain-code-high-level', async () =>
            executeRecipe('explainCodeHighLevel')
        ),
        vscode.commands.registerCommand('cody.recipe.generate-unit-test', async () =>
            executeRecipe('generateUnitTest')
        ),
        vscode.commands.registerCommand('cody.recipe.generate-docstring', async () =>
            executeRecipe('generateDocstring')
        ),
        vscode.commands.registerCommand('cody.recipe.translate-to-language', async () =>
            executeRecipe('translateToLanguage')
        ),
        vscode.commands.registerCommand('cody.recipe.git-history', () => executeRecipe('gitHistory')),

        vscode.window.registerWebviewViewProvider('cody.chat', chatProvider)
    )

    if (config.experimentalSuggest) {
        subscriptions.push(
            vscode.commands.registerCommand('cody.experimental.suggest', async () => {
                await fetchAndShowCompletions(wsCompletionsClient, documentProvider, history)
            })
        )
    }

    setContextActivated(true)

    return vscode.Disposable.from(...subscriptions)
}

const NOOP_DISPOSABLE: vscode.Disposable = {
    dispose: () => {
        /* noop */
    },
}

function setContextActivated(activated: boolean): void {
    vscode.commands.executeCommand('setContext', 'sourcegraph.cody.activated', activated).then(undefined, undefined)
}
