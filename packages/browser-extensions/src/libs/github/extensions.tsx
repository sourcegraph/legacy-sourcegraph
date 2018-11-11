import * as H from 'history'
import { ContributableMenu } from '../../../../extensions-client-common/src/api/protocol'
import { CommandListPopoverButton } from '../../../../extensions-client-common/src/app/CommandList'
import { Controller as ClientController } from '../../../../extensions-client-common/src/client/controller'
import { Controller } from '../../../../extensions-client-common/src/controller'
import { Settings, SettingsSubject } from '../../../../extensions-client-common/src/settings'

import * as React from 'react'
import { render } from 'react-dom'
import { GlobalDebug } from '../../shared/components/GlobalDebug'
import { ShortcutProvider } from '../../shared/components/ShortcutProvider'

export function getCommandPaletteMount(): HTMLElement {
    const headerElem = document.querySelector('div.HeaderMenu>div:last-child')
    if (!headerElem) {
        throw new Error('Unable to find command pallete mount')
    }

    const commandListClass = 'command-palette-button'

    const createCommandList = (): HTMLElement => {
        const commandListElem = document.createElement('div')
        commandListElem.className = commandListClass
        headerElem!.appendChild(commandListElem)

        return commandListElem
    }

    return document.querySelector<HTMLElement>('.' + commandListClass) || createCommandList()
}

export function getGlobalDebugMount(): HTMLElement {
    const globalDebugClass = 'global-debug'

    const createGlobalDebugMount = (): HTMLElement => {
        const globalDebugElem = document.createElement('div')
        globalDebugElem.className = globalDebugClass
        document.body.appendChild(globalDebugElem)

        return globalDebugElem
    }

    return document.querySelector<HTMLElement>('.' + globalDebugClass) || createGlobalDebugMount()
}

// TODO: remove with old inject
export function injectExtensionsGlobalComponents(
    {
        extensionsController,
        extensionsContextController,
    }: {
        extensionsController: ClientController<SettingsSubject, Settings>
        extensionsContextController: Controller<SettingsSubject, Settings>
    },
    location: H.Location
): void {
    render(
        <ShortcutProvider>
            <CommandListPopoverButton
                extensionsController={extensionsController}
                menu={ContributableMenu.CommandPalette}
                extensions={extensionsContextController}
                location={location}
            />
        </ShortcutProvider>,
        getCommandPaletteMount()
    )

    render(<GlobalDebug extensionsController={extensionsController} location={location} />, getGlobalDebugMount())
}
