import { Meta } from '@storybook/react'
import React, { useCallback } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Input } from './Input'

const Story: Meta = {
    title: 'wildcard/Input',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: Input,
        design: {
            type: 'figma',
            name: 'Figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A1943',
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
            <Input title="Input raw" value={selected} onChange={handleChange} />
            <Input
                value={selected}
                title="Input valid"
                onChange={handleChange}
                message="random message"
                status="valid"
                disable={false}
                placeholder="testing this one"
            />
            <Input
                value={selected}
                title="Input loading"
                onChange={handleChange}
                message="random message"
                status="loading"
                placeholder="loading status input"
            />
            <Input
                value={selected}
                title="Input error"
                onChange={handleChange}
                message="a message with error"
                status="error"
                placeholder="error status input"
            />
            <Input
                value={selected}
                title="Disabled input"
                onChange={handleChange}
                message="random message"
                disable={true}
                placeholder="disable status input"
            />

            <Input
                value={selected}
                title="Input small"
                onChange={handleChange}
                message="random message"
                status="valid"
                disable={false}
                placeholder="testing this one"
                size="small"
            />
        </>
    )
}
