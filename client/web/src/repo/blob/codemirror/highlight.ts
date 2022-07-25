import { Extension, RangeSetBuilder, StateEffect, StateField } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { JsonDocument, JsonOccurrence, SyntaxKind } from '../../../lsif/lsif-typed'

/**
 * This data structure combines the syntax highlighting data received from the
 * server with a lineIndex map (implemented as array), for fast lookup by line
 * number, with minimal additional impact on memory (e.g. garbage collection).
 */
interface HighlightIndex {
    occurrences: JsonOccurrence[]
    lineIndex: (number | undefined)[]
}

/**
 * Parses JSON-encoded SCIP syntax highlighting data and creates a line index.
 * NOTE: This assumes that the data is sorted and does not contain overlapping
 * ranges.
 */
function createHighlightTable(json: string): HighlightIndex {
    const lineIndex: (number | undefined)[] = []

    const occurrences = (JSON.parse(json) as JsonDocument).occurrences ?? []
    let previousEndline: number | undefined

    for (let index = 0; index < occurrences.length; index++) {
        let current = occurrences[index]
        const startLine = current.range[0]
        const endLine = current.range.length === 3 ? startLine : current.range[2]

        if (previousEndline !== startLine) {
            lineIndex[startLine] = index
        }

        if (startLine !== endLine) {
            lineIndex[endLine] = index
        }

        previousEndline = endLine
    }

    return { occurrences, lineIndex }
}

/**
 * Set JSON encoded SCIP highlighting data
 */
export const setSCIPData = StateEffect.define<string>()

/**
 * Extension to convert SCIP-encoded highlighting information to decorations.
 * The SCIP data should be set/updated via the `setSCIPData` effect.
 */
export function syntaxHighlight(initialSCIPJSON: string): Extension {
    return StateField.define<HighlightIndex>({
        create: () => createHighlightTable(initialSCIPJSON),

        update(value, transaction) {
            let newSCIPData = ''

            for (const effect of transaction.effects) {
                if (effect.is(setSCIPData)) {
                    newSCIPData = effect.value
                    break
                }
            }

            return newSCIPData ? createHighlightTable(newSCIPData) : value
        },

        provide: field =>
            ViewPlugin.fromClass(
                class {
                    decorationCache: Partial<Record<SyntaxKind, Decoration>> = {}
                    decorations: DecorationSet = Decoration.none

                    constructor(view: EditorView) {
                        this.decorations = this.computeDecorations(view)
                    }

                    update(update: ViewUpdate) {
                        if (update.viewportChanged) {
                            this.decorations = this.computeDecorations(update.view)
                        }
                    }

                    computeDecorations(view: EditorView): DecorationSet {
                        const { from, to } = view.viewport

                        // Determine the start and end lines of the current viewport
                        const fromLine = view.state.doc.lineAt(from)
                        const toLine = view.state.doc.lineAt(to)

                        const { occurrences, lineIndex } = view.state.field(field)

                        // Find index of first relevant token
                        let startIndex: number | undefined = undefined
                        {
                            let line = fromLine.number - 1
                            do {
                                startIndex = lineIndex[line++]
                            } while (startIndex === undefined && line < lineIndex.length)
                        }

                        const builder = new RangeSetBuilder<Decoration>()

                        // Cache current line object
                        let line = fromLine

                        if (startIndex !== undefined) {
                            // Iterate over the rendered line (numbers) and get the
                            // corresponding occurrences from the highlighting table.
                            for (let index = startIndex; index < occurrences.length; index++) {
                                const occurrence = occurrences[index]

                                if (occurrence.range[0] > toLine.number) {
                                    break
                                }

                                if (occurrence.syntaxKind === undefined) {
                                    continue
                                }

                                // Fetch new line information if necessary
                                if (fromLine.number !== occurrence.range[0] + 1) {
                                    line = view.state.doc.line(occurrence.range[0] + 1)
                                }

                                builder.add(
                                    line.from + occurrence.range[1],
                                    occurrence.range.length === 3
                                        ? line.from + occurrence.range[2]
                                        : view.state.doc.line(occurrence.range[2] + 1).from + occurrence.range[3],
                                    this.decorationCache[occurrence.syntaxKind] ||
                                        (this.decorationCache[occurrence.syntaxKind] = Decoration.mark({
                                            class: `hl-typed-${SyntaxKind[occurrence.syntaxKind]}`,
                                        }))
                                )
                            }
                        }

                        return builder.finish()
                    }
                },
                { decorations: plugin => plugin.decorations }
            ),
    })
}
