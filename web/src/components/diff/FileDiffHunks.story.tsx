import { storiesOf } from '@storybook/react'
import { radios, boolean } from '@storybook/addon-knobs'
import React from 'react'
import { FileDiffHunks } from './FileDiffHunks'
import { createMemoryHistory } from 'history'
import webStyles from '../../SourcegraphWebApp.scss'
import { FileDiffHunkFields, DiffHunkLineType } from '../../graphql-operations'

export const DEMO_HUNKS: FileDiffHunkFields[] = [
    {
        oldRange: { lines: 7, startLine: 3 },
        newRange: { lines: 7, startLine: 3 },
        oldNoNewlineAt: false,
        section: 'func awesomeness(param string) (int, error) {',
        highlight: {
            aborted: false,
            lines: [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    v, err := makeAwesome()',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    if err != nil {',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '        fmt.Printf("wow: %v", err)',
                },
                {
                    kind: DiffHunkLineType.DELETED,
                    html: '        return err',
                },
                {
                    kind: DiffHunkLineType.ADDED,
                    html: '        return nil, err',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    }',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    return v.Score, nil',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '}',
                },
            ],
        },
    },
]

const { add } = storiesOf('web/FileDiffHunks', module).addDecorator(story => {
    // TODO find a way to do this globally for all stories and storybook itself.
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <>
            <style>{webStyles}</style>
            <div className="p-3 container">{story()}</div>
        </>
    )
})

add('One diff hunk', () => (
    <FileDiffHunks
        persistLines={boolean('persistLines', false)}
        fileDiffAnchor="abc"
        lineNumbers={boolean('lineNumbers', true)}
        isLightTheme={true}
        hunks={DEMO_HUNKS}
        className="abcdef"
        location={createMemoryHistory().location}
        history={createMemoryHistory()}
    />
))
