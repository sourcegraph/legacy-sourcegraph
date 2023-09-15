import { FC } from 'react'

import { gql, useMutation } from '@sourcegraph/http-client'
import { Button, ErrorAlert, H2, LoadingSpinner, Modal, Text } from '@sourcegraph/wildcard'

import { SearchJobNode } from '../../../graphql-operations'
import { SearchJobCard } from '../SearchJobCard/SearchJobCard'

import styles from './SearchJobModal.module.scss'

const DELETE_SEARCH_JOB = gql`
    mutation DeleteSearchJob($id: ID!) {
        deleteSearchJob(id: $id) {
            alwaysNil
        }
    }
`

interface SearchJobModalProps {
    searchJob: SearchJobNode
    onDismiss: () => void
}

export const SearchJobDeleteModal: FC<SearchJobModalProps> = props => {
    const { searchJob, onDismiss } = props

    const [deleteSearchJob, { loading, error }] = useMutation(DELETE_SEARCH_JOB, {
        onCompleted: (data, clientOptions) => {
            // Delete just deleted search job from the apollo cache
            clientOptions?.client?.cache.evict({ id: searchJob.id })
            onDismiss()
        },
    })

    return (
        <Modal position="center" aria-label="Delete search job" onDismiss={onDismiss}>
            <H2>Do you want to delete search job?</H2>

            <Text className="mt-4">
                <b>Note:</b> All query runs across all repositories will be stopped or canceled in case if the search
                job is still running, search results will be deleted.
            </Text>

            <SearchJobCard searchJob={searchJob} />

            {error && <ErrorAlert error={error} className="mt-3" />}

            <footer className={styles.footer}>
                <Button variant="secondary" outline={true} onClick={onDismiss}>
                    Cancel
                </Button>
                <Button
                    variant="danger"
                    disabled={loading}
                    onClick={() => deleteSearchJob({ variables: { id: searchJob.id } })}
                >
                    {loading ? (
                        <>
                            <LoadingSpinner /> Deleting
                        </>
                    ) : (
                        <>Delete</>
                    )}
                </Button>
            </footer>
        </Modal>
    )
}

const CANCEL_SEARCH_JOB = gql`
    mutation CancelSearchJob($id: ID!) {
        cancelSearchJob(id: $id) {
            alwaysNil
        }
    }
`

const CREATE_SEARCH_JOB = gql`
    mutation CreateSearchJob($query: String!) {
        createSearchJob(query: $query) {
            id
            query
            state
            URL
            startedAt
            finishedAt
            repoStats {
                total
                completed
                failed
                inProgress
            }
            creator {
                id
                displayName
                username
                avatarURL
            }
        }
    }
`

export const RerunSearchJobModal: FC<SearchJobModalProps> = props => {
    const { searchJob, onDismiss } = props

    const [cancelSearchJob, { loading: cancelLoading, error: cancelError }] = useMutation(CANCEL_SEARCH_JOB)
    const [createSearchJob, { loading: creationLoading, error: creationError }] = useMutation(CREATE_SEARCH_JOB)

    const loading = cancelLoading || creationLoading
    const error = cancelError || creationError

    const handleRerunClick = async (): Promise<void> => {
        await cancelSearchJob({ variables: { id: searchJob.id } })
        await createSearchJob({ variables: { query: searchJob.query } })

        onDismiss()
    }

    return (
        <Modal position="center" aria-label="Delete search job" onDismiss={onDismiss}>
            <H2>Do you want to re-run search job?</H2>

            <Text className="mt-4">
                <b>Note:</b> Re-run will produce new search job, the current search job will be cancelled.
            </Text>

            <SearchJobCard searchJob={searchJob} />

            {error && <ErrorAlert error={error} className="mt-3" />}

            <footer className={styles.footer}>
                <Button variant="secondary" outline={true} onClick={onDismiss}>
                    Cancel
                </Button>
                <Button variant="primary" disabled={loading} onClick={() => handleRerunClick()}>
                    {loading ? (
                        <>
                            <LoadingSpinner /> Re-running
                        </>
                    ) : (
                        <>Rerun</>
                    )}
                </Button>
            </footer>
        </Modal>
    )
}

export const CancelSearchJobModal: FC<SearchJobModalProps> = props => {
    const { searchJob, onDismiss } = props

    const [cancelSearchJob, { loading, error }] = useMutation(CANCEL_SEARCH_JOB)

    return (
        <Modal position="center" aria-label="Delete search job" onDismiss={onDismiss}>
            <H2>Do you want to cancel search job?</H2>

            <Text className="mt-4">
                <b>Note:</b> All query runs across all repositories and revisions will be stopped. You can re-run this
                search job later.
            </Text>

            <SearchJobCard searchJob={searchJob} />

            {error && <ErrorAlert error={error} className="mt-3" />}

            <footer className={styles.footer}>
                <Button variant="secondary" outline={true} onClick={onDismiss}>
                    Cancel
                </Button>
                <Button
                    variant="danger"
                    disabled={loading}
                    onClick={() => cancelSearchJob({ variables: { id: searchJob.id } })}
                >
                    {loading ? (
                        <>
                            <LoadingSpinner /> Canceling
                        </>
                    ) : (
                        <>Cancel</>
                    )}
                </Button>
            </footer>
        </Modal>
    )
}
