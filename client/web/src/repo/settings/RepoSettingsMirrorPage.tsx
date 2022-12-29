import React, { useEffect, useState } from 'react'

import { mdiLock } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'
import { RouteComponentProps } from 'react-router'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import {
    Container,
    PageHeader,
    LoadingSpinner,
    FeedbackText,
    Button,
    Link,
    Alert,
    Icon,
    Input,
    Text,
    Code,
    ErrorAlert,
} from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import {
    CheckMirrorRepositoryConnectionResult,
    CheckMirrorRepositoryConnectionVariables,
    RecloneRepositoryResult,
    RecloneRepositoryVariables,
    SettingsAreaRepositoryFields,
    SettingsAreaRepositoryResult,
    SettingsAreaRepositoryVariables,
    UpdateMirrorRepositoryResult,
    UpdateMirrorRepositoryVariables,
} from '../../graphql-operations'
import {
    CHECK_MIRROR_REPOSITORY_CONNECTION,
    RECLONE_REPOSITORY_MUTATION,
    UPDATE_MIRROR_REPOSITORY,
} from '../../site-admin/backend'
import { eventLogger } from '../../tracking/eventLogger'
import { DirectImportRepoAlert } from '../DirectImportRepoAlert'

import { FETCH_SETTINGS_AREA_REPOSITORY_GQL } from './backend'
import { ActionContainer, BaseActionContainer } from './components/ActionContainer'

import styles from './RepoSettingsMirrorPage.module.scss'

interface UpdateMirrorRepositoryActionContainerProps {
    repo: SettingsAreaRepositoryFields
    onDidUpdateRepository: () => Promise<void>
    disabled: boolean
    disabledReason: string | undefined
    history: H.History
}

const UpdateMirrorRepositoryActionContainer: React.FunctionComponent<
    UpdateMirrorRepositoryActionContainerProps
> = props => {
    const [updateRepo] = useMutation<UpdateMirrorRepositoryResult, UpdateMirrorRepositoryVariables>(
        UPDATE_MIRROR_REPOSITORY,
        { variables: { repository: props.repo.id } }
    )

    const run = async (): Promise<void> => {
        await updateRepo()
        await props.onDidUpdateRepository()
    }

    let title: React.ReactNode
    let description: React.ReactNode
    let buttonLabel: React.ReactNode
    let buttonDisabled = false
    let info: React.ReactNode
    if (props.repo.mirrorInfo.cloneInProgress) {
        title = 'Cloning in progress...'
        description =
            <Code>{props.repo.mirrorInfo.cloneProgress}</Code> ||
            'This repository is currently being cloned from its remote repository.'
        buttonLabel = (
            <span>
                <LoadingSpinner /> Cloning...
            </span>
        )
        buttonDisabled = true
        info = <DirectImportRepoAlert className={styles.alert} />
    } else if (props.repo.mirrorInfo.cloned) {
        const updateSchedule = props.repo.mirrorInfo.updateSchedule
        title = (
            <>
                <div>
                    Last refreshed:{' '}
                    {props.repo.mirrorInfo.updatedAt ? <Timestamp date={props.repo.mirrorInfo.updatedAt} /> : 'unknown'}{' '}
                </div>
                {updateSchedule && (
                    <div>
                        Next scheduled update <Timestamp date={updateSchedule.due} /> (position{' '}
                        {updateSchedule.index + 1} out of {updateSchedule.total} in the schedule)
                    </div>
                )}
                {props.repo.mirrorInfo.updateQueue && !props.repo.mirrorInfo.updateQueue.updating && (
                    <div>
                        Queued for update (position {props.repo.mirrorInfo.updateQueue.index + 1} out of{' '}
                        {props.repo.mirrorInfo.updateQueue.total} in the queue)
                    </div>
                )}
            </>
        )
        if (!updateSchedule) {
            description = 'This repository is automatically updated when accessed by a user.'
        } else {
            description =
                'This repository is automatically updated from its remote repository periodically and when accessed by a user.'
        }
        buttonLabel = 'Refresh now'
    } else {
        title = 'Clone this repository'
        description = 'This repository has not yet been cloned from its remote repository.'
        buttonLabel = 'Clone now'
    }

    return (
        <ActionContainer
            title={title}
            description={<div>{description}</div>}
            buttonLabel={buttonLabel}
            buttonDisabled={buttonDisabled || props.disabled}
            buttonSubtitle={props.disabledReason}
            flashText="Added to queue"
            info={info}
            run={run}
            history={props.history}
        />
    )
}

interface CheckMirrorRepositoryConnectionActionContainerProps {
    repo: SettingsAreaRepositoryFields
    onDidUpdateReachability: (reachable: boolean) => void
    history: H.History
}

const CheckMirrorRepositoryConnectionActionContainer: React.FunctionComponent<
    CheckMirrorRepositoryConnectionActionContainerProps
> = props => {
    const [checkConnection, { data, loading, error }] = useMutation<
        CheckMirrorRepositoryConnectionResult,
        CheckMirrorRepositoryConnectionVariables
    >(CHECK_MIRROR_REPOSITORY_CONNECTION, {
        variables: { repository: props.repo.id, name: null },
        onCompleted: result => {
            props.onDidUpdateReachability(result.checkMirrorRepositoryConnection.error === null)
        },
        onError: () => {
            props.onDidUpdateReachability(false)
        },
    })

    useEffect(() => {
        checkConnection().catch(() => {})
    }, [checkConnection])

    return (
        <BaseActionContainer
            title="Check connection to remote repository"
            description={<span>Diagnose problems cloning or updating from the remote repository.</span>}
            action={
                <Button
                    disabled={loading}
                    onClick={() => {
                        checkConnection().catch(() => {})
                    }}
                    variant="primary"
                >
                    Check connection
                </Button>
            }
            details={
                <>
                    {error && <ErrorAlert className={styles.alert} error={error} />}
                    {loading && (
                        <Alert className={classNames('mb-0', styles.alert)} variant="primary">
                            <LoadingSpinner /> Checking connection...
                        </Alert>
                    )}
                    {data &&
                        !loading &&
                        (data.checkMirrorRepositoryConnection.error === null ? (
                            <Alert className={classNames('mb-0', styles.alert)} variant="success">
                                The remote repository is reachable.
                            </Alert>
                        ) : (
                            <Alert className={classNames('mb-0', styles.alert)} variant="danger">
                                <Text>The remote repository is unreachable. Logs follow.</Text>
                                <div>
                                    <pre className={styles.log}>
                                        <Code>{data.checkMirrorRepositoryConnection.error}</Code>
                                    </pre>
                                </div>
                            </Alert>
                        ))}
                </>
            }
            className="mb-0"
        />
    )
}

// Add interface for props then create component
interface CorruptionLogProps extends RouteComponentProps<{}> {
    corruptedAt: any
    logItems: any[]
}

const CorruptionLogsContainer: React.FunctionComponent<CorruptionLogProps> = props => {
    let health = (
        <Alert className={classNames('mb-0', styles.alert)} variant="success">
            The repository is currently not corrupt
        </Alert>
    )
    if (props.corruptedAt) {
        health = (
            <Alert className={classNames('mb-0', styles.alert)} variant="warning">
                The repository is corrupt, check the log entries below for more info and consider recloning.
            </Alert>
        )
    }

    const logEvents = []
    for (let log of props.logItems) {
        logEvents.push(
            <li className="list-group-item px-2 py-2">
                <div className="d-flex flex-column align-items-center justify-content-between">
                    <Text className="overflow-auto text-monospace h-25">{log.reason}</Text>
                    <small className="text-muted mb-0">
                        <Timestamp date={log.timestamp} />
                    </small>
                </div>
            </li>
        )
    }
    // create log item list
    return (
        <BaseActionContainer
            title="Repository corruption"
            description={<span>Recent corruption events that have been detected on this repository.</span>}
            details={
                <div className="flex-1">
                    {health}
                    <br />
                    <ul className="list-group">{logEvents}</ul>
                </div>
            }
        />
    )
}

interface RepoSettingsMirrorPageProps extends RouteComponentProps<{}> {
    repo: SettingsAreaRepositoryFields
    history: H.History
}

/**
 * The repository settings mirror page.
 */
export const RepoSettingsMirrorPage: React.FunctionComponent<
    React.PropsWithChildren<RepoSettingsMirrorPageProps>
> = props => {
    eventLogger.logPageView('RepoSettingsMirror')
    const [reachable, setReachable] = useState<boolean>()
    const [recloneRepository] = useMutation<RecloneRepositoryResult, RecloneRepositoryVariables>(
        RECLONE_REPOSITORY_MUTATION,
        {
            variables: { repo: props.repo.id },
        }
    )

    const { data, error, refetch } = useQuery<SettingsAreaRepositoryResult, SettingsAreaRepositoryVariables>(
        FETCH_SETTINGS_AREA_REPOSITORY_GQL,
        {
            variables: { name: props.repo.name },
            pollInterval: 3000,
        }
    )

    const repo = data?.repository ? data.repository : props.repo

    const onDidUpdateReachability = (reachable: boolean | undefined): void => setReachable(reachable)

    return (
        <>
            <PageTitle title="Mirror settings" />
            <PageHeader path={[{ text: 'Mirroring and cloning' }]} headingElement="h2" className="mb-3" />
            <Container className="repo-settings-mirror-page">
                {error && <ErrorAlert error={error} />}

                <div className="form-group">
                    <Input
                        value={repo.mirrorInfo.remoteURL || '(unknown)'}
                        readOnly={true}
                        className="mb-0"
                        label={
                            <>
                                {' '}
                                Remote repository URL{' '}
                                <small className="text-info">
                                    <Icon aria-hidden={true} svgPath={mdiLock} /> Only visible to site admins
                                </small>
                            </>
                        }
                    />
                    {repo.viewerCanAdminister && (
                        <small className="form-text text-muted">
                            Configure repository mirroring in{' '}
                            <Link to="/site-admin/external-services">external services</Link>.
                        </small>
                    )}
                </div>
                {repo.mirrorInfo.lastError && (
                    <Alert variant="warning">
                        {/* TODO: This should not be a list item, but it was before this was refactored. */}
                        <li className="d-flex w-100">Error updating repo:</li>
                        <li className="d-flex w-100">{repo.mirrorInfo.lastError}</li>
                    </Alert>
                )}
                <UpdateMirrorRepositoryActionContainer
                    repo={repo}
                    onDidUpdateRepository={async () => {
                        await refetch()
                    }}
                    disabled={typeof reachable === 'boolean' && !reachable}
                    disabledReason={typeof reachable === 'boolean' && !reachable ? 'Not reachable' : undefined}
                    history={props.history}
                />
                <ActionContainer
                    title="Reclone repository"
                    description={
                        <div>
                            This will delete the repository from disk and reclone it.
                            <div className="mt-2">
                                <span className="font-weight-bold text-danger">WARNING</span>: This can take a long
                                time, depending on how large the repository is. The repository will be unsearchable
                                while the reclone is in progress.
                            </div>
                        </div>
                    }
                    buttonVariant="danger"
                    buttonLabel={
                        repo.mirrorInfo.cloneInProgress ? (
                            <span>
                                <LoadingSpinner /> Cloning...
                            </span>
                        ) : (
                            'Reclone'
                        )
                    }
                    buttonDisabled={repo.mirrorInfo.cloneInProgress}
                    flashText="Recloning repo"
                    run={async () => {
                        await recloneRepository()
                    }}
                    history={props.history}
                />
                <CheckMirrorRepositoryConnectionActionContainer
                    repo={repo}
                    onDidUpdateReachability={onDidUpdateReachability}
                    history={props.history}
                />
                {reachable === false && (
                    <Alert variant="info">
                        Problems cloning or updating this repository?
                        <ul className={styles.steps}>
                            <li className={styles.step}>
                                Inspect the <strong>Check connection</strong> error log output to see why the remote
                                repository is not reachable.
                            </li>
                            <li className={styles.step}>
                                <Code weight="bold">No ECDSA host key is known ... Host key verification failed?</Code>{' '}
                                See{' '}
                                <Link to="/help/admin/repo/auth#ssh-authentication-config-keys-known-hosts">
                                    SSH repository authentication documentation
                                </Link>{' '}
                                for how to provide an SSH <Code>known_hosts</Code> file with the remote host's SSH host
                                key.
                            </li>
                            <li className={styles.step}>
                                Consult <Link to="/help/admin/repo/add">Sourcegraph repositories documentation</Link>{' '}
                                for resolving other authentication issues (such as HTTPS certificates and SSH keys).
                            </li>
                            <li className={styles.step}>
                                <FeedbackText headerText="Questions?" />
                            </li>
                        </ul>
                    </Alert>
                )}
                <CorruptionLogsContainer
                    corruptedAt={repo.mirrorInfo.corruptedAt}
                    logItems={repo.mirrorInfo.corruptionLogs}
                    history={props.history}
                />
            </Container>
        </>
    )
}
