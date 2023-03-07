import * as vscode from 'vscode'

import { Completion, LLMDebugInfo } from '@sourcegraph/cody-common'

export class CompletionsDocumentProvider implements vscode.TextDocumentContentProvider, vscode.HoverProvider {
	private completionsByUri: {
		[uri: string]: {
			groups: CompletionGroup[]
			status: 'done' | 'notdone'
		}
	} = {}

	private isDebug(): boolean {
		return vscode.workspace.getConfiguration().get<boolean>('cody.debug') === true
	}

	private fireDocumentChanged(uri: vscode.Uri): void {
		this.onDidChangeEmitter.fire(uri)
	}

	public clearCompletions(uri: vscode.Uri): void {
		delete this.completionsByUri[uri.toString()]
		this.fireDocumentChanged(uri)
	}

	public addCompletions(uri: vscode.Uri, lang: string, completions: Completion[], debug?: LLMDebugInfo): void {
		if (!this.completionsByUri[uri.toString()]) {
			this.completionsByUri[uri.toString()] = { groups: [], status: 'notdone' }
		}
		this.completionsByUri[uri.toString()].groups.push({
			lang,
			completions: completions.map(completion => ({
				...completion,
				insertText: `${completion.prefixText}🡆${completion.insertText}`,
			})),
			debug,
		})
		this.fireDocumentChanged(uri)
	}

	public setCompletionsDone(uri: vscode.Uri): void {
		const completions = this.completionsByUri[uri.toString()]
		if (!completions) {
			return
		}
		completions.status = 'done'
		this.fireDocumentChanged(uri)
	}

	public onDidChangeEmitter = new vscode.EventEmitter<vscode.Uri>()
	public onDidChange = this.onDidChangeEmitter.event

	public provideTextDocumentContent(uri: vscode.Uri): string {
		const completionGroups = this.completionsByUri[uri.toString()]
		if (!completionGroups) {
			return 'Loading...'
		}
		return (
			(completionGroups.status === 'notdone' ? 'Loading additional...\n\n' : '') +
			completionGroups.groups
				.map(({ completions, lang }) =>
					completions
						.map((completion, index) => {
							let sectionText = `${headerize(
								`${completion.label} (${index + 1}/${completions.length})`,
								60
							)}`
							if (this.isDebug()) {
								if (completion.finishReason) {
									sectionText += '\n> Finish reason:' + completion.finishReason
								}
							}
							sectionText += `\n\`\`\`${lang}\n${completion.insertText}\n\`\`\``
							return sectionText
						})
						.join('\n\n')
				)
				.join('\n\n')
		)
	}

	public provideHover(document: vscode.TextDocument, position: vscode.Position): vscode.ProviderResult<vscode.Hover> {
		const completionGroups = this.completionsByUri[document.uri.toString()]
		if (!completionGroups) {
			return null
		}

		const wordRange = document.getWordRangeAtPosition(position, /[\w:-]+/)
		if (!wordRange) {
			return null
		}
		const word = document.getText(wordRange)
		for (const { completions, debug } of completionGroups.groups) {
			if (!debug) {
				continue
			}
			for (const { label } of completions) {
				if (label === word) {
					const rawPrompt = new vscode.MarkdownString(`<pre>${debug.prompt}</pre>`)
					rawPrompt.supportHtml = true
					return new vscode.Hover([
						'Options:',
						new vscode.MarkdownString(`\`\`\`\n${JSON.stringify(debug.llmOptions, null, 2)}\n\`\`\``),
						'Prompt:',
						rawPrompt,
					])
				}
			}
		}
		return null
	}
}

export interface CompletionGroup {
	lang: string
	completions: (vscode.InlineCompletionItem & Completion)[]
	debug?: LLMDebugInfo
}

function headerize(label: string, width: number): string {
	const prefix = '# ======= '
	let buffer = width - label.length - prefix.length - 1
	if (buffer < 0) {
		buffer = 0
	}
	return `${prefix}${label} ${'='.repeat(buffer)}`
}
