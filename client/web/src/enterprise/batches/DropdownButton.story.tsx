import { storiesOf } from '@storybook/react'
import React from 'react'

import { EnterpriseWebStory } from '../components/EnterpriseWebStory'

import { Action, DropdownButton } from './DropdownButton'

const { add } = storiesOf('web/batches/DropdownButton', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

// eslint-disable-next-line @typescript-eslint/require-await
const onTrigger = async (onDone: () => void) => onDone()

const action: Action = {
    type: 'action-type',
    buttonLabel: 'Action',
    dropdownTitle: 'Action',
    dropdownDescription: 'Perform an action',
    isAvailable: () => true,
    onTrigger,
}

const disabledAction: Action = {
    type: 'disabled-action-type',
    buttonLabel: 'Disabled action',
    dropdownTitle: 'Disabled action',
    dropdownDescription: 'Perform an action, if only this were enabled',
    isAvailable: () => false,
    onTrigger,
}

const experimentalAction: Action = {
    type: 'experimental-action-type',
    buttonLabel: 'Experimental action',
    dropdownTitle: 'Experimental action',
    dropdownDescription: 'Perform a super cool action that might explode',
    isAvailable: () => true,
    onTrigger,
    experimental: true,
}

add('No actions', () => <EnterpriseWebStory>{() => <DropdownButton actions={[]} />}</EnterpriseWebStory>)

add('Single action', () => <EnterpriseWebStory>{() => <DropdownButton actions={[action]} />}</EnterpriseWebStory>)

add('Multiple actions without default', () => (
    <EnterpriseWebStory>
        {() => <DropdownButton actions={[action, disabledAction, experimentalAction]} />}
    </EnterpriseWebStory>
))

add('Multiple actions with default', () => (
    <EnterpriseWebStory>
        {() => <DropdownButton actions={[action, disabledAction, experimentalAction]} defaultAction={1} />}
    </EnterpriseWebStory>
))
