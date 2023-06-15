import { LRUCache } from 'lru-cache'
import * as vscode from 'vscode'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'

import { debug } from '../log'
import { CodyStatusBar } from '../services/StatusBar'

import { CompletionsCache } from './cache'
import { getContext } from './context'
import { getCurrentDocContext } from './document'
import { History } from './history'
import * as CompletionLogger from './logger'
import { detectMultilineMode } from './multiline'
import { postProcess } from './post-process'
import { Provider, ProviderConfig } from './providers/provider'
import { SNIPPET_WINDOW_SIZE } from './utils'

export const inlineCompletionsCache = new CompletionsCache()

export class CodyCompletionItemProvider implements vscode.InlineCompletionItemProvider {
    private promptChars: number
    private maxPrefixChars: number
    private maxSuffixChars: number
    private abortOpenInlineCompletions: () => void = () => {}
    private lastContentChanges: LRUCache<string, 'add' | 'del'> = new LRUCache<string, 'add' | 'del'>({
        max: 10,
    })

    constructor(
        private providerConfig: ProviderConfig,
        private history: History,
        private statusBar: CodyStatusBar,
        private codebaseContext: CodebaseContext,
        private responsePercentage = 0.1,
        private prefixPercentage = 0.6,
        private suffixPercentage = 0.1,
        private disableTimeouts = false
    ) {
        this.promptChars =
            providerConfig.maximumContextCharacters - providerConfig.maximumContextCharacters * responsePercentage
        this.maxPrefixChars = Math.floor(this.promptChars * this.prefixPercentage)
        this.maxSuffixChars = Math.floor(this.promptChars * this.suffixPercentage)

        debug('CodyCompletionProvider:initialized', `provider: ${providerConfig.identifier}`)

        vscode.workspace.onDidChangeTextDocument(event => {
            const document = event.document
            const changes = event.contentChanges

            if (changes.length <= 0) {
                return
            }

            const text = changes[0].text
            this.lastContentChanges.set(document.fileName, text.length > 0 ? 'add' : 'del')
        })
    }

    public async provideInlineCompletionItems(
        document: vscode.TextDocument,
        position: vscode.Position,
        context: vscode.InlineCompletionContext,
        // Making it optional here to execute multiple suggestion in parallel from the CLI script.
        token?: vscode.CancellationToken
    ): Promise<vscode.InlineCompletionItem[]> {
        try {
            return await this.provideInlineCompletionItemsInner(document, position, context, token)
        } catch (error) {
            if (isAbortError(error)) {
                return []
            }

            console.error(error)
            debug('CodyCompletionProvider:inline:error', `${error.toString()}\n${error.stack}`)
            return []
        }
    }

    private async provideInlineCompletionItemsInner(
        document: vscode.TextDocument,
        position: vscode.Position,
        context: vscode.InlineCompletionContext,
        token?: vscode.CancellationToken
    ): Promise<vscode.InlineCompletionItem[]> {
        const abortController = new AbortController()
        if (token) {
            this.abortOpenInlineCompletions()
            token.onCancellationRequested(() => abortController.abort())
            this.abortOpenInlineCompletions = () => abortController.abort()
        }

        CompletionLogger.clear()

        const currentEditor = vscode.window.activeTextEditor
        if (!currentEditor || currentEditor?.document.uri.scheme === 'cody') {
            return []
        }

        const docContext = getCurrentDocContext(document, position, this.maxPrefixChars, this.maxSuffixChars)
        if (!docContext) {
            return []
        }

        const { prefix, suffix, prevLine: sameLinePrefix, prevNonEmptyLine } = docContext
        const sameLineSuffix = suffix.slice(0, suffix.indexOf('\n'))

        // Avoid showing completions when we're deleting code (Cody can only insert code at the
        // moment)
        const lastChange = this.lastContentChanges.get(document.fileName) ?? 'add'
        if (lastChange === 'del') {
            // When a line was deleted, only look up cached items and only include them if the
            // untruncated prefix matches. This fixes some weird issues where the completion would
            // render if you insert whitespace but not on the original place when you delete it
            // again
            const cachedCompletions = inlineCompletionsCache.get(prefix, false)
            if (cachedCompletions?.isExactPrefix) {
                return toInlineCompletionItems(cachedCompletions.logId, cachedCompletions.completions)
            }
            return []
        }

        const cachedCompletions = inlineCompletionsCache.get(prefix)
        if (cachedCompletions) {
            return toInlineCompletionItems(cachedCompletions.logId, cachedCompletions.completions)
        }

        const similarCode = await getContext({
            currentEditor,
            prefix,
            suffix,
            history: this.history,
            jaccardDistanceWindowSize: SNIPPET_WINDOW_SIZE,
            maxChars: this.promptChars,
            codebaseContext: this.codebaseContext,
        })

        const completers: Provider[] = []
        let timeout: number
        let multilineMode: null | 'block' = null
        // VS Code does not show completions if we are in the process of writing a word or if a
        // selected completion info is present (so something is selected from the completions
        // dropdown list based on the lang server) and the returned completion range does not
        // contain the same selection.
        if (context.selectedCompletionInfo || /[A-Za-z]$/.test(sameLinePrefix)) {
            return []
        }
        // If we have a suffix in the same line as the cursor and the suffix contains any word
        // characters, do not attempt to make a completion. This means we only make completions if
        // we have a suffix in the same line for special characters like `)]}` etc.
        //
        // VS Code will attempt to merge the remainder of the current line by characters but for
        // words this will easily get very confusing.
        if (/\w/.test(sameLineSuffix)) {
            return []
        }
        // In this case, VS Code won't be showing suggestions anyway and we are more likely to want
        // suggested method names from the language server instead.
        if (context.triggerKind === vscode.InlineCompletionTriggerKind.Invoke || sameLinePrefix.endsWith('.')) {
            return []
        }

        const sharedProviderOptions = {
            prefix,
            suffix,
            fileName: document.fileName,
            languageId: document.languageId,
            snippets: similarCode,
            responsePercentage: this.responsePercentage,
            prefixPercentage: this.prefixPercentage,
            suffixPercentage: this.suffixPercentage,
        }

        if (
            (multilineMode = detectMultilineMode(
                prefix,
                prevNonEmptyLine,
                sameLinePrefix,
                sameLineSuffix,
                document.languageId
            ))
        ) {
            timeout = 200
            completers.push(
                this.providerConfig.create({
                    ...sharedProviderOptions,
                    n: 3,
                    multilineMode,
                })
            )
        } else if (sameLinePrefix.trim() === '') {
            // The current line is empty
            timeout = 20
            completers.push(
                this.providerConfig.create({
                    ...sharedProviderOptions,
                    n: 3,
                    multilineMode: null,
                })
            )
        } else {
            // The current line has a suffix
            timeout = 200
            completers.push(
                this.providerConfig.create({
                    ...sharedProviderOptions,
                    n: 3,
                    multilineMode: null,
                })
            )
        }

        // We don't need to make a request at all if the signal is already aborted after the
        // debounce
        if (abortController.signal.aborted) {
            return []
        }

        const logId = CompletionLogger.start({
            type: 'inline',
            multilineMode,
            providerIdentifier: this.providerConfig.identifier,
        })
        const stopLoading = this.statusBar.startLoading('Completions are being generated')

        // Overwrite the abort handler to also update the loading state
        const previousAbort = this.abortOpenInlineCompletions
        this.abortOpenInlineCompletions = () => {
            previousAbort()
            stopLoading()
        }

        if (!this.disableTimeouts) {
            await new Promise<void>(resolve => setTimeout(resolve, timeout))
        }

        const completions = (
            await Promise.all(completers.map(c => c.generateCompletions(abortController.signal)))
        ).flat()

        // Post process
        const processedCompletions = completions.map(completion =>
            postProcess({
                prefix,
                suffix,
                multiline: multilineMode !== null,
                languageId: document.languageId,
                completion,
            })
        )

        // Filter results
        const visibleResults = filterCompletions(processedCompletions)

        // Rank results
        const rankedResults = rankCompletions(visibleResults)

        stopLoading()

        if (rankedResults.length > 0) {
            CompletionLogger.suggest(logId)
            inlineCompletionsCache.add(logId, rankedResults)
            return toInlineCompletionItems(logId, rankedResults)
        }

        CompletionLogger.noResponse(logId)
        return []
    }
}

export interface Completion {
    prefix: string
    content: string
    stopReason?: string
}

function toInlineCompletionItems(logId: string, completions: Completion[]): vscode.InlineCompletionItem[] {
    return completions.map(
        completion =>
            new vscode.InlineCompletionItem(completion.content, undefined, {
                title: 'Completion accepted',
                command: 'cody.completions.inline.accepted',
                arguments: [{ codyLogId: logId }],
            })
    )
}

function rankCompletions(completions: Completion[]): Completion[] {
    // TODO(philipp-spiess): Improve ranking to something more complex then just length
    return completions.sort((a, b) => b.content.split('\n').length - a.content.split('\n').length)
}

function filterCompletions(completions: Completion[]): Completion[] {
    return completions.filter(c => c.content.trim() !== '')
}

function isAbortError(error: Error): boolean {
    return (
        // http module
        error.message === 'aborted' ||
        // fetch
        error.message.includes('The operation was aborted') ||
        error.message.includes('The user aborted a request')
    )
}
