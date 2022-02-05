import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { DotcomGettingStartedPage } from './DotcomGettingStartedPage'

const { add } = storiesOf('web/batches/DotcomGettingStartedPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
            disableSnapshot: false,
        },
    })

add('Overview', () => <WebStory>{() => <DotcomGettingStartedPage />}</WebStory>)
