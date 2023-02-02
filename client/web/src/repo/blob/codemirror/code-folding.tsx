import { foldEffect, foldGutter, foldKeymap, foldService } from '@codemirror/language'
import { EditorState, Extension, Line, StateField } from '@codemirror/state'
import { EditorView, keymap, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { mdiMenuDown, mdiMenuRight } from '@mdi/js'
import { createRoot } from 'react-dom/client'

import { Icon } from '@sourcegraph/wildcard/src'

import { rangeToCmSelection } from './occurrence-utils'
import { getCodeIntelTooltipState } from './token-selection/code-intel-tooltips'

enum CharCode {
    /**
     * The `\t` character.
     */
    Tab = 9,

    Space = 32,
}

/**
 * Returns:
 *  - -1 => the line consists of whitespace
 *  - otherwise => the indent level is returned value
 */
function computeIndentLevel(line: string, tabSize: number): number {
    let indent = 0
    let index = 0
    const len = line.length

    while (index < len) {
        const charCode = line.charCodeAt(index)

        if (charCode === CharCode.Space) {
            indent++
        } else if (charCode === CharCode.Tab) {
            indent = indent - (indent % tabSize) + tabSize
        } else {
            break
        }
        index++
    }

    if (index === len) {
        return -1 // line only consists of whitespace
    }

    return indent
}

/**
 * Computes foldable ranges based on lines indentation.
 *
 * Implements similar to [VSCode indent-based folding](https://sourcegraph.com/github.com/microsoft/vscode@e3d73a5a2fd03412d83887a073c9c71bad38e964/-/blob/src/vs/editor/contrib/folding/browser/indentRangeProvider.ts?L126-200) logic.
 */
function computeFoldableRanges(state: EditorState): [Line, Line][] {
    const ranges: [Line, Line][] = []
    const previousRanges = [{ indent: -1, endAbove: state.doc.lines + 1 }]

    for (let lineNumber = state.doc.lines; lineNumber > 0; lineNumber--) {
        const line = state.doc.line(lineNumber)
        const indent = computeIndentLevel(line.text, state.tabSize)
        if (indent === -1) {
            continue
        }

        let previous = previousRanges[previousRanges.length - 1]
        if (previous.indent > indent) {
            // remove ranges with larger indent
            do {
                previousRanges.pop()
                previous = previousRanges[previousRanges.length - 1]
            } while (previous.indent > indent)

            // new folding range
            const endLineNumber = previous.endAbove - 1
            if (endLineNumber - lineNumber >= 1) {
                // should be at least 2 lines
                ranges.push([line, state.doc.line(endLineNumber)])
            }
        }
        if (previous.indent === indent) {
            previous.endAbove = lineNumber
        } else {
            // previous.indent < indent
            // new range with a bigger indent
            previousRanges.push({ indent, endAbove: lineNumber })
        }
    }

    return ranges
}

/**
 * Stores foldable lines ranges.
 *
 * Value is computed when field is initialized and never updated.
 */
const foldingRanges = StateField.define<[Line, Line][]>({
    create: computeFoldableRanges,
    update(value) {
        return value
    },
})

function getFoldRange(state: EditorState, lineStart: number): { from: number; to: number } | null {
    const ranges = state.field(foldingRanges)

    const range = ranges.find(([start]) => start.number === state.doc.lineAt(lineStart).number)

    if (!range) {
        return null
    }

    const [start, end] = range

    return { from: start.to, to: end.to }
}

/**
 * Enables indent-based code folding.
 */
export function codeFoldingExtension(): Extension {
    return [
        foldingRanges,

        foldService.of(getFoldRange),

        foldGutter({
            markerDOM(open: boolean): HTMLElement {
                const container = document.createElement('div')
                const root = createRoot(container)
                root.render(
                    <Icon aria-hidden={true} svgPath={open ? mdiMenuDown : mdiMenuRight} className="fold-icon" />
                )
                return container
            },
        }),

        keymap.of(foldKeymap),

        ViewPlugin.define(view => ({
            update(update: ViewUpdate) {
                for (const transaction of update.transactions) {
                    for (const effect of transaction.effects) {
                        if (effect.is(foldEffect)) {
                            const focusedOccurrence = getCodeIntelTooltipState(view, 'focus')?.occurrence
                            if (focusedOccurrence) {
                                const range = rangeToCmSelection(view.state, focusedOccurrence.range)
                                if (range.from >= effect.value.from && range.to <= effect.value.to) {
                                    // Occurrence is inside the folded range.
                                    // It will be removed from DOM triggering editor's blur.
                                    // Focus it back for the keyboard navigation to work.
                                    view.contentDOM.focus()
                                }
                            }
                        }
                    }
                }
            },
        })),

        EditorView.theme({
            '.cm-foldGutter': {
                height: '1rem',
                width: '1rem',
            },
            '.cm-foldGutter .fold-icon': {
                width: '100%',
                height: '100%',
                color: 'var(--text-muted)',
                cursor: 'pointer',
            },
            '.cm-foldPlaceholder': {
                background: 'var(--color-bg-3)',
                borderColor: 'var(--border-color)',
            },
        }),
    ]
}
