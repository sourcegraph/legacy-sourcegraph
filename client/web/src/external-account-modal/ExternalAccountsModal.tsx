import { useEffect, useState } from 'react'

import type { AuthenticatedUser } from 'src/auth'
import { BrandLogo } from 'src/components/branding/BrandLogo'
import type { UserExternalAccountsWithAccountDataVariables } from 'src/graphql-operations'
import type { SourcegraphContext } from 'src/jscontext'
import { ExternalAccountsSignIn } from 'src/user/settings/auth/ExternalAccountsSignIn'
import type { UserExternalAccount, UserExternalAccountsResult } from 'src/user/settings/auth/UserSettingsSecurityPage'
import { USER_EXTERNAL_ACCOUNTS } from 'src/user/settings/backend'

import type { ErrorLike } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { Button, ErrorAlert, H2, LoadingSpinner, Modal, Text } from '@sourcegraph/wildcard'

import styles from './ExternalAccountsModal.module.scss'

export interface ExternalAccountsModalProps {
    authenticatedUser: AuthenticatedUser
    isVisible: boolean
    setVisible: (visible: boolean) => void
    onDismiss: () => void
    context: Pick<SourcegraphContext, 'authProviders'>
}

export const ExternalAccountsModal: React.FunctionComponent<ExternalAccountsModalProps> = props => {
    const [accounts, setAccounts] = useState<{ fetched?: UserExternalAccount[]; lastRemoved?: string }>({
        fetched: [],
        lastRemoved: '',
    })

    const { data, loading, refetch } = useQuery<
        UserExternalAccountsResult,
        UserExternalAccountsWithAccountDataVariables
    >(USER_EXTERNAL_ACCOUNTS, {
        variables: { username: props.authenticatedUser.username },
    })

    const [error, setError] = useState<ErrorLike>()

    const handleError = (error: ErrorLike): [] => {
        setError(error)
        return []
    }

    useEffect(() => {
        setAccounts({ fetched: data?.user?.externalAccounts.nodes, lastRemoved: '' })
    }, [data])

    const onAccountRemoval = (removeId: string, name: string): void => {
        // keep every account that doesn't match removeId
        setAccounts({ fetched: accounts.fetched?.filter(({ id }) => id !== removeId), lastRemoved: name })
    }

    const onAccountAdd = (): void => {
        refetch({ username: props.authenticatedUser.username })
            .then(() => {})
            .catch(handleError)
    }

    return (
        <Modal
            aria-label="Connect your external accounts"
            isOpen={props.isVisible}
            onDismiss={props.onDismiss}
            className={styles.modal}
            position="center"
        >
            <div className={styles.title}>
                <BrandLogo variant="symbol" isLightTheme={true} />
                <div>
                    <H2>Sourcegraph setup: permissions & security</H2>
                    <Text>Connect external identity providers to access private repositories.</Text>
                </div>
            </div>
            <hr />
            {loading && <LoadingSpinner />}
            {error && <ErrorAlert className="mb-3" error={error} />}
            {accounts.fetched && (
                <ExternalAccountsSignIn
                    onDidAdd={onAccountAdd}
                    onDidError={handleError}
                    onDidRemove={onAccountRemoval}
                    accounts={accounts.fetched}
                    authProviders={props.context.authProviders}
                />
            )}
            <hr />
            <Button onClick={props.onDismiss} className={styles.skip} size="lg" variant="secondary" outline={true}>
                Done
            </Button>
        </Modal>
    )
}
