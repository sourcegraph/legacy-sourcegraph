import { Facet, RangeSetBuilder } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, PluginValue, ViewUpdate, ViewPlugin } from '@codemirror/view'
import classNames from 'classnames'

import { logger } from '@sourcegraph/common'
import { SyntaxKind } from '@sourcegraph/shared/src/codeintel/scip'

import { BlobInfo } from '../CodeMirrorBlob'

import { syntaxHighlight } from './highlight'

import styles from './links.module.scss'

const SUPPORTED_KINDS = new Set<SyntaxKind>([SyntaxKind.Comment, SyntaxKind.StringLiteral])

const LINK_REGEX =
    /(?:(?:https?|ftp|file):\/\/|www\.|ftp\.)(?:\([\w!#$%&+,./:=?@|~-]*\)|[\w!#$%&+,./:=?@|~-])*(?:\([\w!#$%&+,./:=?@|~-]*\)|[\w#$%&+/=@|~])/gim

/**
 * Iterates through highlighting occurrences and match any URLs in `SUPPORTED_KINDS`.
 * Converts matches into <a> tags with a permanent underline and a relevant href.
 */
class LinkBuilder implements PluginValue {
    // private decorationCache: Partial<Record<SyntaxKind, Decoration>> = {}
    public decorations: DecorationSet = Decoration.none

    constructor(view: EditorView) {
        this.decorations = this.computeDecorations(view)
    }

    public update(update: ViewUpdate): void {
        if (update.viewportChanged) {
            this.decorations = this.computeDecorations(update.view)
        }
    }

    private computeDecorations(view: EditorView): DecorationSet {
        const builder = new RangeSetBuilder<Decoration>()
        try {
            const { from, to } = view.viewport

            // Determine the start and end lines of the current viewport
            const fromLine = view.state.doc.lineAt(from)
            const toLine = view.state.doc.lineAt(to)

            const { occurrences, lineIndex } = view.state.facet(syntaxHighlight)

            // Find index of first relevant token
            let startIndex: number | undefined
            {
                let line = fromLine.number - 1
                do {
                    startIndex = lineIndex[line++]
                } while (startIndex === undefined && line < lineIndex.length)
            }

            if (startIndex !== undefined) {
                // Iterate over the rendered line (numbers) and get the
                // corresponding occurrences from the highlighting table.
                const textDocument = view.state.doc

                for (let index = startIndex; index < occurrences.length; index++) {
                    const occurrence = occurrences[index]
                    const occurrenceStartLine = occurrence.range.start.line + 1

                    // Skip if out of viewport
                    if (occurrenceStartLine > toLine.number) {
                        break
                    }

                    // Skip if the syntax kind is not supported.
                    if (occurrence.kind === undefined || !SUPPORTED_KINDS.has(occurrence.kind)) {
                        continue
                    }

                    // Skip if the range is not on a single line.
                    // We know that we will not match a link here.
                    if (!occurrence.range.isSingleLine()) {
                        continue
                    }

                    const line = textDocument.line(occurrenceStartLine)
                    const matches = line.text.matchAll(LINK_REGEX)

                    for (const match of matches) {
                        if (match.index) {
                            const from = Math.min(line.from + match.index, line.to)
                            const to = Math.min(line.from + match.index + match[0].length, line.to)

                            const decoration = Decoration.mark({
                                tagName: 'a',
                                class: classNames(styles.link, `hl-typed-${SyntaxKind[occurrence.kind]}`),
                                attributes: {
                                    href: match[0],
                                },
                            })

                            builder.add(from, to, decoration)
                        }
                    }
                }
            }
        } catch (error) {
            logger.error('Failed to build links', error)
        }
        return builder.finish()
    }
}

export const buildLinks = Facet.define<BlobInfo>({
    static: true,
    compareInput: (blobInfoA, blobInfoB) => blobInfoA.lsif === blobInfoB.lsif,
    enables: ViewPlugin.fromClass(LinkBuilder, { decorations: plugin => plugin.decorations }),
})
