import * as anthropic from '@anthropic-ai/sdk'

import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import {
    CodeCompletionParameters,
    CodeCompletionResponse,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { Completion } from '.'
import { ReferenceSnippet } from './context'
import { Message, messagesToText } from './prompts'

export abstract class CompletionProvider {
    constructor(
        protected completionsClient: SourcegraphNodeCompletionsClient,
        protected promptChars: number,
        protected responseTokens: number,
        protected snippets: ReferenceSnippet[],
        protected prefix: string,
        protected injectPrefix: string,
        protected defaultN: number = 1
    ) {}

    // Returns the content specific prompt excluding additional referenceSnippets
    protected abstract createPromptPrefix(): Message[]

    public emptyPromptLength(): number {
        const promptNoSnippets = messagesToText(this.createPromptPrefix())
        return promptNoSnippets.length - 10 // extra 10 chars of buffer cuz who knows
    }

    // Creates the resulting prompt and adds as many snippets from the reference
    // list as possible.
    protected createPrompt(): string {
        const prefixMessages = this.createPromptPrefix()
        const referenceSnippetMessages: Message[] = []

        let remainingChars = this.promptChars - this.emptyPromptLength()
        for (const snippet of this.snippets) {
            const snippetMessages: Message[] = [
                {
                    role: 'human',
                    text:
                        `Add the following code snippet (from file ${snippet.filename}) to your knowledge base:\n` +
                        '```' +
                        `\n${snippet.text}\n` +
                        '```',
                },
                {
                    role: 'ai',
                    text: 'Okay, I have added it to my knowledge base.',
                },
            ]
            const numSnippetChars = messagesToText(snippetMessages).length + 1
            if (numSnippetChars > remainingChars) {
                break
            }
            referenceSnippetMessages.push(...snippetMessages)
            remainingChars -= numSnippetChars
        }

        return messagesToText([...referenceSnippetMessages, ...prefixMessages])
    }

    public abstract generateCompletions(abortSignal: AbortSignal, n?: number): Promise<Completion[]>
}

export class MultilineCompletionProvider extends CompletionProvider {
    protected createPromptPrefix(): Message[] {
        // TODO(beyang): escape 'Human:' and 'Assistant:'
        const prefix = this.prefix.trim()

        const prefixLines = prefix.split('\n')
        if (prefixLines.length === 0) {
            throw new Error('no prefix lines')
        }

        let prefixMessages: Message[]
        if (prefixLines.length > 2) {
            const endLine = Math.max(Math.floor(prefixLines.length / 2), prefixLines.length - 5)
            prefixMessages = [
                {
                    role: 'human',
                    text:
                        'Complete the following file:\n' +
                        '```' +
                        `\n${prefixLines.slice(0, endLine).join('\n')}\n` +
                        '```',
                },
                {
                    role: 'ai',
                    text: `Here is the completion of the file:\n\`\`\`\n${prefixLines.slice(endLine).join('\n')}`,
                },
            ]
        } else {
            prefixMessages = [
                {
                    role: 'human',
                    text: 'Write some code',
                },
                {
                    role: 'ai',
                    text: `Here is some code:\n\`\`\`\n${prefix}`,
                },
            ]
        }

        return prefixMessages
    }

    private postProcess(completion: string): string {
        const endBlockIndex = completion.indexOf('```')
        if (endBlockIndex !== -1) {
            return completion.slice(0, endBlockIndex).trimEnd()
        }
        return completion.trimEnd()
    }

    public async generateCompletions(abortSignal: AbortSignal, n?: number): Promise<Completion[]> {
        const prefix = this.prefix.trim()

        // Create prompt
        const prompt = this.createPrompt()
        if (prompt.length > this.promptChars) {
            throw new Error('prompt length exceeded maximum alloted chars')
        }

        // Issue request
        const responses = await batchCompletions(
            this.completionsClient,
            {
                prompt,
                stopSequences: [anthropic.HUMAN_PROMPT],
                maxTokensToSample: this.responseTokens,
                model: 'claude-instant-v1.0',
                temperature: 1, // default value (source: https://console.anthropic.com/docs/api/reference)
                topK: -1, // default value
                topP: -1, // default value
            },
            n || this.defaultN,
            abortSignal
        )
        // Post-process
        return responses.map(resp => ({
            prefix,
            prompt,
            content: this.postProcess(resp.completion),
            stopReason: resp.stopReason,
        }))
    }
}

export class EndOfLineCompletionProvider extends CompletionProvider {
    protected createPromptPrefix(): Message[] {
        // TODO(beyang): escape 'Human:' and 'Assistant:'
        const prefixLines = this.prefix.split('\n')
        if (prefixLines.length === 0) {
            throw new Error('no prefix lines')
        }

        let prefixMessages: Message[]
        if (prefixLines.length > 2) {
            const endLine = Math.max(Math.floor(prefixLines.length / 2), prefixLines.length - 5)
            prefixMessages = [
                {
                    role: 'human',
                    text:
                        'Complete the following file:\n' +
                        '```' +
                        `\n${prefixLines.slice(0, endLine).join('\n')}\n` +
                        '```',
                },
                {
                    role: 'ai',
                    text:
                        'Here is the completion of the file:\n' +
                        '```' +
                        `\n${prefixLines.slice(endLine).join('\n')}${this.injectPrefix}`,
                },
            ]
        } else {
            prefixMessages = [
                {
                    role: 'human',
                    text: 'Write some code',
                },
                {
                    role: 'ai',
                    text: `Here is some code:\n\`\`\`\n${this.prefix}${this.injectPrefix}`,
                },
            ]
        }

        return prefixMessages
    }

    private postProcess(completion: string): string {
        // Sometimes Claude emits an extra space
        if (
            completion.length > 0 &&
            completion.startsWith(' ') &&
            this.prefix.length > 0 &&
            this.prefix.endsWith(' ')
        ) {
            completion = completion.slice(1)
        }
        // Insert the injected prefix back in
        if (this.injectPrefix.length > 0) {
            completion = this.injectPrefix + completion
        }
        // Strip out trailing markdown block and trim trailing whitespace
        const endBlockIndex = completion.indexOf('```')
        if (endBlockIndex !== -1) {
            return completion.slice(0, endBlockIndex).trimEnd()
        }
        return completion.trimEnd()
    }

    public async generateCompletions(abortSignal: AbortSignal, n?: number): Promise<Completion[]> {
        const prefix = this.prefix + this.injectPrefix

        // Create prompt
        const prompt = this.createPrompt()
        if (prompt.length > this.promptChars) {
            throw new Error('prompt length exceeded maximum alloted chars')
        }

        // Issue request
        const responses = await batchCompletions(
            this.completionsClient,
            {
                prompt,
                stopSequences: [anthropic.HUMAN_PROMPT, '\n'],
                maxTokensToSample: this.responseTokens,
                model: 'claude-instant-v1.0',
                temperature: 1,
                topK: -1,
                topP: -1,
            },
            n || this.defaultN,
            abortSignal
        )
        // Post-process
        return responses.map(resp => ({
            prefix,
            prompt,
            content: this.postProcess(resp.completion),
            stopReason: resp.stopReason,
        }))
    }
}

async function batchCompletions(
    client: SourcegraphNodeCompletionsClient,
    params: CodeCompletionParameters,
    n: number,
    abortSignal: AbortSignal
): Promise<CodeCompletionResponse[]> {
    const responses: Promise<CodeCompletionResponse>[] = []
    for (let i = 0; i < n; i++) {
        responses.push(client.complete(params, abortSignal))
    }
    return Promise.all(responses)
}
