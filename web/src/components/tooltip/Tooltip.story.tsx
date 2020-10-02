import { storiesOf } from '@storybook/react'
import React, { useCallback } from 'react'
import { WebStory } from '../WebStory'
import { Tooltip } from './Tooltip'

const { add } = storiesOf('web/Tooltip', module).addDecorator(story => <WebStory>{() => story()}</WebStory>)

add(
    'Hover',
    () => (
        <>
            <Tooltip />
            <p>
                You can <strong data-tooltip="Tooltip 1">hover me</strong> or{' '}
                <strong data-tooltip="Tooltip 2">me</strong>.
            </p>
        </>
    ),
    {
        chromatic: {
            disable: true,
        },
    }
)

const PinnedTooltip: React.FunctionComponent = () => {
    const clickElement = useCallback((element: HTMLElement | null) => {
        if (element) {
            element.click()
        }
    }, [])
    return (
        <>
            <Tooltip />
            <span data-tooltip="My tooltip" ref={clickElement}>
                Example
            </span>
            <p>
                <small>
                    (A pinned tooltip is shown when the target element is rendered, without any user interaction
                    needed.)
                </small>
            </p>
        </>
    )
}
add('Pinned', () => <PinnedTooltip />)
