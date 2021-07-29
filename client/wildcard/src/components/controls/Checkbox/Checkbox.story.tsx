import { Meta } from '@storybook/react'
import React, { useCallback } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Checkbox } from './Checkbox'

const Story: Meta = {
    title: 'wildcard/Checkbox',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: Checkbox,
        design: {
            type: 'figma',
            name: 'Figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A1353',
        },
    },
}

// eslint-disable-next-line import/no-default-export
export default Story

export const Simple = () => {
    const [selected, setSelected] = React.useState('')

    const handleChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setSelected(event.target.value)
    }, [])

    return (
        <>
            <Checkbox
                value="first"
                checked={selected === 'first'}
                onChange={handleChange}
                label="First"
                message="Hello world!"
            />
            <Checkbox
                value="second"
                checked={selected === 'second'}
                onChange={handleChange}
                label="Second"
                message="Hello world!"
            />
            <Checkbox
                value="third"
                checked={selected === 'third'}
                onChange={handleChange}
                label="Third"
                message="Hello world!"
            />
        </>
    )
}
