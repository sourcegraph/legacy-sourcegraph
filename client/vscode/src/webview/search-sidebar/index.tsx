import '../platform/polyfills'

import { ShortcutProvider } from '@slimsag/react-shortcuts'
import { VSCodeProgressRing } from '@vscode/webview-ui-toolkit/react'
import * as Comlink from 'comlink'
import React, { useMemo } from 'react'
import { render } from 'react-dom'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { AnchorLink, setLinkComponent, useObservable, WildcardThemeContext, Tooltip } from '@sourcegraph/wildcard'

import { ExtensionCoreAPI } from '../../contract'
import { createEndpointsForWebToNode } from '../comlink/webviewEndpoint'
import { createPlatformContext, WebviewPageProps } from '../platform/context'
import { adaptSourcegraphThemeToEditorTheme } from '../theming/sourcegraphTheme'

import { createSearchSidebarAPI } from './api'
import { AuthSidebarView } from './AuthSidebarView'
import { ContextInvalidatedSidebarView } from './ContextInvalidatedSidebarView'
import { HistoryHomeSidebar } from './HistorySidebarView'
import { SearchSidebarView } from './SearchSidebarView'

const vsCodeApi = window.acquireVsCodeApi()

const { proxy, expose } = createEndpointsForWebToNode(vsCodeApi)

export const extensionCoreAPI: Comlink.Remote<ExtensionCoreAPI> = Comlink.wrap(proxy)

const platformContext = createPlatformContext(extensionCoreAPI)

const searchSidebarAPI = createSearchSidebarAPI({
    platformContext,
    instanceURL: document.documentElement.dataset.instanceUrl!,
})
Comlink.expose(searchSidebarAPI, expose)

setLinkComponent(AnchorLink)

const themes = adaptSourcegraphThemeToEditorTheme()

const Main: React.FC = () => {
    // TODO: make sure we only rerender on necessary changes
    const state = useObservable(useMemo(() => wrapRemoteObservable(extensionCoreAPI.observeState()), []))

    const authenticatedUser = useObservable(
        useMemo(() => wrapRemoteObservable(extensionCoreAPI.getAuthenticatedUser()), [])
    )

    const instanceURL = useObservable(useMemo(() => wrapRemoteObservable(extensionCoreAPI.getInstanceURL()), []))

    const theme = useObservable(themes)

    const settingsCascade = useObservable(
        useMemo(() => wrapRemoteObservable(extensionCoreAPI.observeSourcegraphSettings()), [])
    )
    // Do not block rendering on settings unless we observe UI jitter

    // TODO: If init is taking too long, show a message.
    // Also check if anything has errored out.

    // If any of the remote values have yet to load.
    const initialized =
        state !== undefined &&
        authenticatedUser !== undefined &&
        instanceURL !== undefined &&
        theme !== undefined &&
        settingsCascade !== undefined
    if (!initialized) {
        return <VSCodeProgressRing />
    }

    const webviewPageProps: WebviewPageProps = {
        extensionCoreAPI,
        platformContext,
        theme,
        authenticatedUser,
        settingsCascade,
        instanceURL,
    }

    if (state.status === 'context-invalidated') {
        return <ContextInvalidatedSidebarView {...webviewPageProps} />
    }

    // If a search hasn't been performed yet
    if (state.status === 'search-home' || !state.context.submittedSearchQueryState) {
        // TODO: should we hide the access token form permanently if an unauthenticated user
        // has performed a search before? Or just for this session?
        if (!authenticatedUser) {
            return <AuthSidebarView {...webviewPageProps} stateStatus={state.status} />
        }
        return <HistoryHomeSidebar {...webviewPageProps} authenticatedUser={authenticatedUser} />
    }

    // <SearchSidebarView> is wrapped w/ React.memo so pass only necessary props.
    return (
        <>
            <SearchSidebarView
                platformContext={platformContext}
                extensionCoreAPI={extensionCoreAPI}
                settingsCascade={settingsCascade}
            />
            {!authenticatedUser && <AuthSidebarView {...webviewPageProps} stateStatus={state.status} />}
        </>
    )
}

render(
    <ShortcutProvider>
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <Main />
            <Tooltip key={1} className="sourcegraph-tooltip" />
        </WildcardThemeContext.Provider>
    </ShortcutProvider>,
    document.querySelector('#root')
)
