import { Icons, IconButton } from '@storybook/components'
import React, { ReactElement } from 'react'

import { useRedesignSubject, REDESIGN_CLASS_NAME } from '@sourcegraph/shared/src/util/useRedesignSubject'

const toggleRedesignClass = (element: HTMLElement, isRedesignEnabled: boolean): void => {
    element.classList.toggle(REDESIGN_CLASS_NAME, !isRedesignEnabled)
}

const updatePreview = (isRedesignEnabled: boolean): void => {
    const iframe = document.querySelector('#storybook-preview-iframe') as HTMLIFrameElement | undefined

    const iframeDocument = iframe?.contentDocument || iframe?.contentWindow?.document
    const body = iframeDocument?.body

    if (body) {
        toggleRedesignClass(body, isRedesignEnabled)
    }
}

const updateManager = (isRedesignEnabled: boolean): void => {
    const manager = document.querySelector('body')

    if (manager) {
        toggleRedesignClass(manager, isRedesignEnabled)
    }
}

export const RedesignToggleStorybook = (): ReactElement => {
    const [redesignSubject, isRedesignEnabled] = useRedesignSubject()

    const handleRedesignToggle = (): void => {
        redesignSubject.next(!isRedesignEnabled)
        updatePreview(!!isRedesignEnabled)
        updateManager(!!isRedesignEnabled)
    }

    return (
        <IconButton
            key="redesign-toolbar"
            active={isRedesignEnabled}
            title={isRedesignEnabled ? 'Disable redesign theme' : 'Enable redesign theme'}
            // eslint-disable-next-line react/jsx-no-bind
            onClick={handleRedesignToggle}
        >
            <Icons icon="beaker" />
        </IconButton>
    )
}
