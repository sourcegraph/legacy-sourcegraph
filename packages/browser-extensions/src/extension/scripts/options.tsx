// We want to polyfill first.
// prettier-ignore
import '../../config/polyfill'

import * as React from 'react'
import { render } from 'react-dom'
import { Subscription } from 'rxjs'
import storage from '../../browser/storage'
import { featureFlagDefaults, FeatureFlags } from '../../browser/types'
import { OptionsContainer, OptionsContainerProps } from '../../libs/options/OptionsContainer'
import { getAccessToken, setAccessToken } from '../../shared/auth/access_token'
import { createAccessToken, fetchAccessTokenIDs } from '../../shared/backend/auth'
import { fetchCurrentUser, fetchSite } from '../../shared/backend/server'
import { OptionsDashboard } from '../../shared/components/options/OptionsDashboard'
import { featureFlags } from '../../shared/util/featureFlags'
import { assertEnv } from '../envAssertion'

assertEnv('OPTIONS')

type State = Pick<FeatureFlags, 'useExtensions'> & { sourcegraphURL: string | null }

const keyIsFeatureFlag = (key: string): key is keyof FeatureFlags =>
    !!Object.keys(featureFlagDefaults).find(k => key === k)

const toggleFeatureFlag = (key: string) => {
    if (keyIsFeatureFlag(key)) {
        featureFlags.toggle(key)
    }
}

class Options extends React.Component<{}, State> {
    public state: State = { sourcegraphURL: null, useExtensions: false }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            storage.observeSync('featureFlags').subscribe(({ useExtensions }) => {
                this.setState({ useExtensions })
            })
        )

        this.subscriptions.add(
            storage.observeSync('sourcegraphURL').subscribe(sourcegraphURL => {
                this.setState({ sourcegraphURL })
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        if (this.state.sourcegraphURL === null) {
            return null
        }

        const props: OptionsContainerProps = {
            sourcegraphURL: this.state.sourcegraphURL,

            ensureValidSite: fetchSite,
            fetchCurrentUser,

            setSourcegraphURL: (url: string) => {
                storage.setSync({ sourcegraphURL: url })
            },

            createAccessToken,
            getAccessToken,
            setAccessToken,
            fetchAccessTokenIDs,

            toggleFeatureFlag,
            featureFlags: [{ key: 'useExtensions', value: this.state.useExtensions }],
        }

        return <OptionsContainer {...props} />
    }
}

const inject = async () => {
    const injectDOM = document.createElement('div')
    injectDOM.id = 'sourcegraph-options-menu'
    injectDOM.className = 'options'
    document.body.appendChild(injectDOM)

    if (await featureFlags.isEnabled('simpleOptionsMenu')) {
        render(<Options />, injectDOM)
    } else {
        storage.getSync(() => {
            render(<OptionsDashboard />, injectDOM)
        })
    }
}

document.addEventListener('DOMContentLoaded', async () => {
    await inject()
})
