import React from 'react'

import { Args, useMemo } from '@storybook/addons'
import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { PopoverTrigger, H1 } from '../..'
import { Button } from '../../Button'

import { FeedbackPrompt } from '.'

import styles from './FeedbackPrompt.module.scss'

const config: Meta = {
    title: 'wildcard/FeedbackPrompt',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],
    parameters: {
        /**
         * Uncomment this once Storybook is upgraded to v6.4.* and the `play` function
         * is used to show the feedback prompt component.
         *
         * https://www.chromatic.com/docs/hoverfocus#javascript-triggered-hover-states
         */
        // chromatic: { disableSnapshot: false },
        component: FeedbackPrompt,
        design: {
            type: 'figma',
            name: 'figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A1',
        },
    },
    argTypes: {
        authenticatedUser: {
            control: { type: 'boolean' },
            defaultValue: true,
        },
    },
}

export default config

const handleSuccessSubmit = () =>
    Promise.resolve({
        errorMessage: undefined,
        isHappinessFeedback: true,
    })
const handleErrorSubmit = () =>
    Promise.resolve({
        errorMessage: 'Something went really wrong',
        isHappinessFeedback: false,
    })

const commonProps = (
    props: Args
): Pick<React.ComponentProps<typeof FeedbackPrompt>, 'authenticatedUser' | 'openByDefault'> => ({
    authenticatedUser: props.authenticatedUser ? { username: 'logan', email: 'logan@example.com' } : null,
    openByDefault: true, // to save storybook viewers from needing to click to see the prompt
})

export const FeedbackPromptWithSuccessResponse: Story = args => (
    <>
        <H1>This is a feedbackPrompt with success response</H1>
        <FeedbackPrompt onSubmit={handleSuccessSubmit} {...useMemo(() => commonProps(args), [args])}>
            <PopoverTrigger
                className={styles.feedbackPrompt}
                as={Button}
                aria-label="Feedback"
                variant="secondary"
                outline={true}
                size="sm"
            >
                <span>Feedback</span>
            </PopoverTrigger>
        </FeedbackPrompt>
    </>
)

export const FeedbackPromptWithErrorResponse: Story = args => (
    <>
        <H1>This is a feedbackPrompt with error response</H1>
        <FeedbackPrompt onSubmit={handleErrorSubmit} {...useMemo(() => commonProps(args), [args])}>
            <PopoverTrigger
                className={styles.feedbackPrompt}
                as={Button}
                aria-label="Feedback"
                variant="secondary"
                outline={true}
                size="sm"
            >
                <span>Feedback</span>
            </PopoverTrigger>
        </FeedbackPrompt>
    </>
)

export const FeedbackPromptWithInModal: Story = args => (
    <>
        <H1>This is a feedbackPrompt in modal</H1>
        <FeedbackPrompt onSubmit={handleSuccessSubmit} modal={true} {...useMemo(() => commonProps(args), [args])}>
            {({ onClick }) => (
                <Button onClick={onClick} aria-label="Feedback" variant="secondary" outline={true} size="sm">
                    <small>Feedback</small>
                </Button>
            )}
        </FeedbackPrompt>
    </>
)
