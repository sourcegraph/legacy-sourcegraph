import classNames from 'classnames'
import React, { useMemo, useState } from 'react'
import { Form } from 'reactstrap'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'
import { currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'
import { Alert } from '@sourcegraph/wildcard'

import { WebviewPageProps } from '../platform/context'

import styles from './AuthSidebarView.module.scss'

interface AuthSidebarViewProps extends WebviewPageProps {
    stateStatus: string
}

/**
 * Rendered by sidebar in search-home state when user doesn't have a valid access token.
 */
export const AuthSidebarView: React.FunctionComponent<AuthSidebarViewProps> = ({
    instanceURL,
    extensionCoreAPI,
    platformContext,
    stateStatus,
}) => {
    const [state, setState] = useState<'initial' | 'validating' | 'success' | 'failure'>('initial')

    const [hasAccount, setHasAccount] = useState(false)

    const signUpURL = useMemo(() => new URL('sign-up?editor=vscode', instanceURL).href, [instanceURL])
    const instanceHostname = useMemo(() => new URL(instanceURL).hostname, [instanceURL])

    const ctaButtonClassName = 'btn btn-primary font-weight-normal w-100 my-1 border-0'
    const buttonLinkClassName = 'btn btn-sm btn-link d-block pl-0'

    const validateAccessToken: React.FormEventHandler<HTMLFormElement> = (event): void => {
        event.preventDefault()
        if (state !== 'validating') {
            const newAccessToken = (event.currentTarget.elements.namedItem('token') as HTMLInputElement).value

            setState('validating')
            const currentAuthStateResult = platformContext
                .requestGraphQL<CurrentAuthStateResult, CurrentAuthStateVariables>({
                    request: currentAuthStateQuery,
                    variables: {},
                    mightContainPrivateInfo: true,
                    overrideAccessToken: newAccessToken,
                })
                .toPromise()

            currentAuthStateResult
                .then(({ data }) => {
                    if (data?.currentUser) {
                        setState('success')
                        return extensionCoreAPI.setAccessToken(newAccessToken)
                    }
                    setState('failure')
                    return
                })
                // v2/debt: Disambiguate network vs auth errors like we do in the browser extension.
                .catch(() => setState('failure'))
        }
        // If successful, update setting. This form will no longer be rendered
    }

    const onSignUpClick = (): void => {
        setHasAccount(true)
        extensionCoreAPI
            .openLink(signUpURL)
            .then(() => {})
            .catch(() => {})

        platformContext.telemetryService.log('VSCESidebarCreateAccount')
    }

    const onLinkClick = (type: 'Sourcegraph' | 'Extension'): void =>
        platformContext.telemetryService.log(`VSCESidebarLearn${type}Click`)

    if (state === 'success') {
        // This form should no longer be rendered as the extension context
        // will be invalidated. We should show a notification that the accessToken
        // has successfully been updated.
        return null
    }

    const renderCommon = (content: JSX.Element): JSX.Element => (
        <div className={classNames(styles.ctaContainer)}>
            {stateStatus === 'search-home' && (
                <div>
                    <h5 className="mt-3 mb-2">Welcome!</h5>
                    <p className={classNames(styles.ctaParagraph)}>
                        The Sourcegraph extension allows you to search millions of open source repositories without
                        cloning them to your local machine.
                    </p>
                    <p className={classNames(styles.ctaParagraph)}>
                        Developers at some of the world’s best software companies use Sourcegraph to onboard to new code
                        bases, find examples, research errors, and resolve incidents.
                    </p>
                    <div>
                        <p className="mb-0">Learn more:</p>
                        <a href="http://sourcegraph.com/" className="my-0" onClick={() => onLinkClick('Sourcegraph')}>
                            Sourcegraph.com
                        </a>
                        <br />
                        <a
                            href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph"
                            className="my-0"
                            onClick={() => onLinkClick('Extension')}
                        >
                            Sourcegraph VS Code extension
                        </a>
                    </div>
                </div>
            )}
            <Form onSubmit={validateAccessToken} className={styles.formContainer}>
                <h5 className="mb-2">Search your private code</h5>
                {content}
            </Form>
        </div>
    )

    if (!hasAccount) {
        return renderCommon(
            <>
                <p className={classNames(styles.ctaParagraph)}>
                    Create an account to enhance search across your private repositories: search multiple repos & commit
                    history, monitor, save searches and more.
                </p>
                <button type="button" onClick={onSignUpClick} className={ctaButtonClassName}>
                    Create an account
                </button>
                <button type="button" onClick={() => setHasAccount(true)} className={buttonLinkClassName}>
                    Have an account?
                </button>
            </>
        )
    }

    return renderCommon(
        <>
            <p className={classNames(styles.ctaParagraph)}>
                Sign in by entering an access token created through your user settings on {instanceHostname}.
            </p>
            <p className={classNames(styles.ctaParagraph)}>
                See our{' '}
                <a
                    href="https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token"
                    onClick={() => platformContext.telemetryService.log('VSCESidebarCreateToken')}
                >
                    user docs
                </a>{' '}
                for a video guide on how to create an access token.
            </p>
            {state === 'failure' && (
                <Alert variant="danger">
                    Unable to verify your access token for {instanceHostname}. Please try again with a new access token.
                </Alert>
            )}
            <LoaderInput loading={state === 'validating'}>
                <input
                    className="input form-control mb-1"
                    type="text"
                    name="token"
                    required={true}
                    autoFocus={true}
                    spellCheck={false}
                    disabled={state === 'validating'}
                    placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                />
            </LoaderInput>
            <button type="submit" disabled={state === 'validating'} className={ctaButtonClassName}>
                Enter access token
            </button>
            <button type="button" onClick={onSignUpClick} className={buttonLinkClassName}>
                Create an account
            </button>
        </>
    )
}
