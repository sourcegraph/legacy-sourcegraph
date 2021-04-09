import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { DecorationMapByLine, decorationStyleForTheme } from '@sourcegraph/shared/src/api/extension/api/decorations'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isDefined, property } from '@sourcegraph/shared/src/util/types'
import * as React from 'react'
import { useLocation } from 'react-router'
import { DiffHunkLineType, FileDiffHunkFields } from '../../graphql-operations'
import { addLineNumberToHunks } from './addLineNumberToHunks'
import { DiffBoundary } from './DiffBoundary'
import { EmptyLine, Line } from './Lines'
import { useSplitDiff } from './useSplitDiff'

export interface DiffHunkProps extends ThemeProps {
    /** The anchor (URL hash link) of the file diff. The component creates sub-anchors with this prefix. */
    fileDiffAnchor: string
    hunk: FileDiffHunkFields
    lineNumbers: boolean
    decorations: Record<'head' | 'base', DecorationMapByLine>
    /**
     * Reflect selected line in url
     *
     * @default true
     */
    persistLines?: boolean
}

// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
const addDecorations = (isLightTheme: boolean, decorationsForLine: TextDocumentDecoration[]) => {
    const lineStyle = decorationsForLine
        .filter(decoration => decoration.isWholeLine)
        .map(decoration => decorationStyleForTheme(decoration, isLightTheme))
        .reduce((style, decoration) => ({ ...style, ...decoration }), {})

    const decorationsWithAfterProperty = decorationsForLine.filter(property('after', isDefined))

    return { lineStyle, decorationsWithAfterProperty }
}

export interface Hunk {
    kind: DiffHunkLineType
    html: string
    anchor: string
    oldLine?: number
    newLine?: number
}

export const DiffSplitHunk: React.FunctionComponent<DiffHunkProps> = ({
    fileDiffAnchor,
    decorations,
    hunk,
    lineNumbers,
    persistLines = true,
    isLightTheme,
}) => {
    const location = useLocation()

    const { hunksWithLineNumber } = addLineNumberToHunks(
        hunk.highlight.lines,
        hunk.newRange.startLine,
        hunk.oldRange.startLine,
        fileDiffAnchor
    )

    const { diff } = useSplitDiff(hunksWithLineNumber)

    const groupHunks = React.useCallback(
        (hunks: Hunk[]): JSX.Element[] => {
            const elements = []
            for (let index = 0; index < hunks.length; index++) {
                const current = hunks[index]

                const lineNumber = (elements[index + 1] ? current.oldLine : current.newLine) as number
                const active = location.hash === `#${current.anchor}`

                const decorationsForLine = [
                    // If the line was deleted, look for decorations in the base revision
                    ...((current.kind === DiffHunkLineType.DELETED && decorations.base.get(lineNumber)) || []),
                    // If the line wasn't deleted, look for decorations in the head revision
                    ...((current.kind !== DiffHunkLineType.DELETED && decorations.head.get(lineNumber)) || []),
                ]

                const { lineStyle, decorationsWithAfterProperty } = addDecorations(isLightTheme, decorationsForLine)

                const lineProps = {
                    persistLines,
                    lineStyle,
                    decorations: decorationsWithAfterProperty,
                    className: active ? 'diff-hunk__line--active' : '',
                    lineNumbers,
                    html: current.html,
                    anchor: current.anchor,
                    kind: current.kind,
                }

                if (current.kind === DiffHunkLineType.UNCHANGED) {
                    // UNCHANGED is displayed on both side
                    elements.push(
                        <tr key={current.anchor} data-testid={current.anchor}>
                            <Line
                                {...lineProps}
                                key={`L${current.anchor}`}
                                id={`L${current.anchor}`}
                                lineNumber={current.oldLine}
                            />
                            <Line
                                {...lineProps}
                                key={`R${current.anchor}`}
                                id={`R${current.anchor}`}
                                lineNumber={current.newLine}
                            />
                        </tr>
                    )
                } else if (current.kind === DiffHunkLineType.DELETED) {
                    const next = hunks[index + 1]
                    // If an ADDED change is following a DELETED change, they should be displayed side by side
                    if (next?.kind === DiffHunkLineType.ADDED) {
                        index = index + 1
                        elements.push(
                            <tr key={current.anchor} data-testid={current.anchor}>
                                <Line {...lineProps} key={current.anchor} lineNumber={current.oldLine} />
                                <Line
                                    {...lineProps}
                                    key={next.anchor}
                                    kind={next.kind}
                                    lineNumber={next.newLine}
                                    anchor={next.anchor}
                                    html={next.html}
                                    className={location.hash === `#${next.anchor}` ? 'diff-hunk__line--active' : ''}
                                />
                            </tr>
                        )
                    } else {
                        // DELETED is following by an empty line
                        elements.push(
                            <tr key={current.anchor} data-testid={current.anchor}>
                                <Line
                                    {...lineProps}
                                    key={current.anchor}
                                    lineNumber={
                                        current.kind === DiffHunkLineType.DELETED ? current.oldLine : lineNumber
                                    }
                                />
                                <EmptyLine />
                            </tr>
                        )
                    }
                } else {
                    // ADDED is preceded by an empty line
                    elements.push(
                        <tr key={current.anchor} data-testid={current.anchor}>
                            <EmptyLine />
                            <Line {...lineProps} key={current.anchor} lineNumber={lineNumber} />
                        </tr>
                    )
                }
            }

            return elements
        },
        [decorations.base, decorations.head, isLightTheme, location.hash, lineNumbers, persistLines]
    )

    const diffView = React.useMemo(() => groupHunks(diff), [diff, groupHunks])

    return (
        <>
            <DiffBoundary {...hunk} contentClassName="diff-hunk__content" lineNumbers={lineNumbers} diffMode="split" />
            {diffView}
        </>
    )
}
