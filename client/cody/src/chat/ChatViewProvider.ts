import path from 'path'

import * as vscode from 'vscode'

import { BotResponseMultiplexer } from '@sourcegraph/cody-shared/src/chat/bot-response-multiplexer'
import { ChatClient } from '@sourcegraph/cody-shared/src/chat/chat'
import { getPreamble } from '@sourcegraph/cody-shared/src/chat/preamble'
import { getRecipe } from '@sourcegraph/cody-shared/src/chat/recipes'
import { Transcript } from '@sourcegraph/cody-shared/src/chat/transcript'
import { ChatMessage, ChatHistory } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { reformatBotMessage } from '@sourcegraph/cody-shared/src/chat/viewHelpers'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { highlightTokens } from '@sourcegraph/cody-shared/src/hallucinations-detector'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'
import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import { version as packageVersion } from '../../package.json'
import { LocalStorage } from '../command/LocalStorageProvider'
import { updateConfiguration } from '../configuration'
import { VSCodeEditor } from '../editor/vscode-editor'
import { logEvent } from '../event-logger'
import { configureExternalServices } from '../external-services'
import { sanitizeServerEndpoint } from '../sanitize'
import { CODY_ACCESS_TOKEN_SECRET, getAccessToken, SecretStorage } from '../secret-storage'
import { TestSupport } from '../test-support'

async function isValidLogin(
    serverEndpoint: string,
    accessToken: string,
    customHeaders: Record<string, string>
): Promise<boolean> {
    const client = new SourcegraphGraphQLAPIClient(sanitizeServerEndpoint(serverEndpoint), accessToken, customHeaders)
    const userId = await client.getCurrentUserId()
    return !isError(userId)
}

export class ChatViewProvider implements vscode.WebviewViewProvider {
    private isMessageInProgress = false
    private cancelCompletionCallback: (() => void) | null = null
    private webview?: vscode.Webview

    private tosVersion = packageVersion

    private currentChatID = ''
    private inputHistory: string[] = []
    private chatHistory: ChatHistory = {}

    // Allows recipes to hook up subscribers to process sub-streams of bot output
    private multiplexer: BotResponseMultiplexer = new BotResponseMultiplexer()

    constructor(
        private extensionPath: string,
        private codebase: string,
        // eslint-disable-next-line
        private serverEndpoint: string,
        private transcript: Transcript,
        private chat: ChatClient,
        private intentDetector: IntentDetector,
        private codebaseContext: CodebaseContext,
        private editor: Editor,
        private secretStorage: SecretStorage,
        private contextType: 'embeddings' | 'keyword' | 'none' | 'blended',
        private rgPath: string,
        private mode: 'development' | 'production',
        private localStorage: LocalStorage,
        private customHeaders: Record<string, string>
    ) {
        if (TestSupport.instance) {
            TestSupport.instance.chatViewProvider.set(this)
        }
        // chat id is used to identify chat session
        this.createNewChatID()
    }

    public static create(
        extensionPath: string,
        codebase: string,
        serverEndpoint: string,
        contextType: 'embeddings' | 'keyword' | 'none' | 'blended',
        secretStorage: SecretStorage,
        localStorage: LocalStorage,
        editor: VSCodeEditor,
        rgPath: string,
        mode: 'development' | 'production',
        intentDetector: IntentDetector,
        codebaseContext: CodebaseContext,
        chatClient: ChatClient,
        customHeaders: Record<string, string>
    ): ChatViewProvider {
        return new ChatViewProvider(
            extensionPath,
            codebase,
            serverEndpoint,
            new Transcript(),
            chatClient,
            intentDetector,
            codebaseContext,
            editor,
            secretStorage,
            contextType,
            rgPath,
            mode,
            localStorage,
            customHeaders
        )
    }

    private async onDidReceiveMessage(message: any): Promise<void> {
        switch (message.command) {
            case 'initialized':
                await this.sendToken()
                this.sendTranscript()
                this.sendChatHistory()
                break
            case 'reset':
                this.onResetChat()
                this.sendChatHistory()
                break
            case 'submit':
                await this.onHumanMessageSubmitted(message.text)
                break
            case 'executeRecipe':
                await this.executeRecipe(message.recipe)
                break
            case 'acceptTOS':
                await this.acceptTOS(message.version)
                break
            case 'settings': {
                const isValid = await isValidLogin(message.serverEndpoint, message.accessToken, this.customHeaders)
                if (isValid) {
                    await updateConfiguration('serverEndpoint', message.serverEndpoint)
                    await this.secretStorage.store(CODY_ACCESS_TOKEN_SECRET, message.accessToken)
                    logEvent('CodyVSCodeExtension:login:clicked')
                }
                this.sendLogin(isValid)
                break
            }
            case 'removeToken':
                await this.secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
                logEvent('CodyVSCodeExtension:codyDeleteAccessToken:clicked')
                break
            case 'removeHistory':
                await this.localStorage.removeChatHistory()
                break
            case 'links':
                await vscode.env.openExternal(vscode.Uri.parse(message.value))
                break
            case 'openFile': {
                const rootPath = this.editor.getWorkspaceRootPath()
                if (rootPath !== null) {
                    const uri = vscode.Uri.file(path.join(rootPath, message.filePath))
                    // This opens the file in the active column.
                    try {
                        const doc = await vscode.workspace.openTextDocument(uri)
                        await vscode.window.showTextDocument(doc)
                    } catch (error) {
                        console.error(`Could not open file: ${error}`)
                    }
                } else {
                    console.error('Could not open file because rootPath is null')
                }
                break
            }
            default:
                console.error('Invalid request type from Webview')
        }
    }

    private async acceptTOS(version: string): Promise<void> {
        this.tosVersion = version
        await vscode.commands.executeCommand('cody.accept-tos', version)
        logEvent('CodyVSCodeExtension:acceptTerms:clicked')
    }

    private createNewChatID(): void {
        this.currentChatID = new Date(Date.now()).toUTCString()
    }

    private sendPrompt(promptMessages: Message[], responsePrefix = ''): void {
        this.cancelCompletion()

        let text = ''

        this.multiplexer.sub(BotResponseMultiplexer.DEFAULT_TOPIC, {
            onResponse: (content: string) => {
                text += content
                this.transcript.addAssistantResponse(reformatBotMessage(text, responsePrefix))
                this.sendTranscript()
                return Promise.resolve()
            },
            onTurnComplete: async () => {
                const lastInteraction = this.transcript.getLastInteraction()
                if (lastInteraction) {
                    const { text, displayText } = lastInteraction.getAssistantMessage()
                    const { text: highlightedDisplayText } = await highlightTokens(displayText, fileExists)
                    this.transcript.addAssistantResponse(text, highlightedDisplayText)
                }
                this.isMessageInProgress = false
                this.cancelCompletionCallback = null
                this.sendTranscript()
                await this.saveChatHistory()
            },
        })

        let textConsumed = 0

        this.cancelCompletionCallback = this.chat.chat(promptMessages, {
            onChange: text => {
                // TODO(dpc): The multiplexer can handle incremental text. Change chat to provide incremental text.
                text = text.slice(textConsumed)
                textConsumed += text.length
                return this.multiplexer.publish(text)
            },
            onComplete: () => this.multiplexer.notifyTurnComplete(),
            onError: err => {
                void vscode.window.showErrorMessage(err)
            },
        })
    }

    private cancelCompletion(): void {
        this.cancelCompletionCallback?.()
        this.cancelCompletionCallback = null
    }

    private onResetChat(): void {
        this.createNewChatID()
        this.cancelCompletion()
        this.isMessageInProgress = false
        this.transcript.reset()
        this.sendTranscript()
    }

    private async onHumanMessageSubmitted(text: string): Promise<void> {
        this.inputHistory.push(text)
        await this.executeRecipe('chat-question', text)
    }

    public async executeRecipe(recipeId: string, humanChatInput: string = ''): Promise<void> {
        if (this.isMessageInProgress) {
            await vscode.window.showErrorMessage(
                'Cannot execute multiple recipes. Please wait for the current recipe to finish.'
            )
        }
        const recipe = getRecipe(recipeId)
        if (!recipe) {
            return
        }

        // Create a new multiplexer to drop any old subscribers
        this.multiplexer = new BotResponseMultiplexer()

        const interaction = await recipe.getInteraction(humanChatInput, {
            editor: this.editor,
            intentDetector: this.intentDetector,
            codebaseContext: this.codebaseContext,
            responseMultiplexer: this.multiplexer,
        })
        if (!interaction) {
            return
        }
        this.isMessageInProgress = true
        this.transcript.addInteraction(interaction)

        this.showTab('chat')
        this.sendTranscript()

        const prompt = await this.transcript.toPrompt(getPreamble(this.codebase))
        this.sendPrompt(prompt, interaction.getAssistantMessage().prefix ?? '')

        logEvent(`CodyVSCodeExtension:recipe:${recipe.getID()}:executed`)
    }

    private showTab(tab: string): void {
        void vscode.commands.executeCommand('cody.chat.focus')
        void this.webview?.postMessage({ type: 'showTab', tab })
    }

    private sendTranscript(): void {
        void this.webview?.postMessage({
            type: 'transcript',
            messages: this.transcript.toChat(),
            isMessageInProgress: this.isMessageInProgress,
        })
    }

    private sendLogin(isValid: boolean): void {
        void this.webview?.postMessage({ type: 'login', isValid })
    }

    /**
     * Sends access token to webview
     */
    private async sendToken(): Promise<void> {
        void this.webview?.postMessage({
            type: 'token',
            value: await getAccessToken(this.secretStorage),
            mode: this.mode,
        })
    }
    /**
     * Save chat history
     */
    private async saveChatHistory(): Promise<void> {
        if (this.transcript) {
            this.chatHistory[this.currentChatID] = this.transcript.toChat()
            const userHistory = {
                chat: this.chatHistory,
                input: this.inputHistory,
            }
            await this.localStorage.setChatHistory(userHistory)
        }
    }
    /**
     * Sends chat history to webview
     */
    private sendChatHistory(): void {
        const localHistory = this.localStorage.getChatHistory()
        if (localHistory) {
            this.chatHistory = localHistory.chat
            this.inputHistory = localHistory.input
        }
        void this.webview?.postMessage({
            type: 'history',
            messages: localHistory,
        })
    }
    /**
     * create webview resources
     */
    public async resolveWebviewView(
        webviewView: vscode.WebviewView,
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        _context: vscode.WebviewViewResolveContext<unknown>,
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        _token: vscode.CancellationToken
    ): Promise<void> {
        this.webview = webviewView.webview

        const extensionPath = vscode.Uri.file(this.extensionPath)
        const webviewPath = vscode.Uri.joinPath(extensionPath, 'dist')

        webviewView.webview.options = {
            enableScripts: true,
            localResourceRoots: [webviewPath],
        }

        // Create Webview
        const root = vscode.Uri.joinPath(webviewPath, 'index.html')
        const bytes = await vscode.workspace.fs.readFile(root)
        const decoded = new TextDecoder('utf-8').decode(bytes)
        const resources = webviewView.webview.asWebviewUri(webviewPath)
        const nonce = this.getNonce()

        webviewView.webview.html = decoded
            .replaceAll('./', `${resources.toString()}/`)
            .replace('/nonce/', nonce)
            .replace('/tos-version/', this.tosVersion.toString())
        webviewView.webview.onDidReceiveMessage(message => this.onDidReceiveMessage(message))
    }

    public transcriptForTesting(testing: TestSupport): ChatMessage[] {
        if (!testing) {
            console.error('used ForTesting method without test support object')
            return []
        }
        return this.transcript.toChat()
    }

    // TODO(beyang): maybe move this into CommandsProvider (should maybe change that to a top-level controller class)
    public async onConfigChange(change: string, codebase: string, serverEndpoint: string): Promise<void> {
        switch (change) {
            case 'token':
            case 'endpoint': {
                const { intentDetector, codebaseContext, chatClient } = await configureExternalServices(
                    serverEndpoint,
                    codebase,
                    this.rgPath,
                    this.editor,
                    this.secretStorage,
                    this.contextType,
                    this.mode,
                    this.customHeaders
                )

                this.codebase = codebase
                this.serverEndpoint = serverEndpoint
                this.intentDetector = intentDetector
                this.codebaseContext = codebaseContext
                this.chat = chatClient

                const action = await vscode.window.showInformationMessage(
                    'Cody configuration has been updated.',
                    'Reload Window'
                )

                logEvent('CodyVSCodeExtension:updateEndpoint:clicked')
                if (action === 'Reload Window') {
                    await vscode.commands.executeCommand('workbench.action.reloadWindow')
                }
                break
            }
        }
    }

    private getNonce(): string {
        let text = ''
        const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
        for (let i = 0; i < 32; i++) {
            text += possible.charAt(Math.floor(Math.random() * possible.length))
        }
        return text
    }
}

function trimPrefix(text: string, prefix: string): string {
    if (text.startsWith(prefix)) {
        return text.slice(prefix.length)
    }
    return text
}

function trimSuffix(text: string, suffix: string): string {
    if (text.endsWith(suffix)) {
        return text.slice(0, -suffix.length)
    }
    return text
}

async function fileExists(filePath: string): Promise<boolean> {
    const patterns = [filePath, '**/' + trimSuffix(trimPrefix(filePath, '/'), '/') + '/**']
    if (!filePath.endsWith('/')) {
        patterns.push('**/' + trimPrefix(filePath, '/') + '*')
    }
    for (const pattern of patterns) {
        const files = await vscode.workspace.findFiles(pattern, null, 1)
        if (files.length > 0) {
            return true
        }
    }
    return false
}
