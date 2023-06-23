import { FC, useEffect } from 'react'

import { noop } from 'lodash'

import { gql, useQuery } from '@sourcegraph/http-client'
import { LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { RepositoriesProgressResult } from '../../../../graphql-operations'

import styles from './AppSetupProgressBar.module.scss'

const REPO_UPLOADING_PROGRESS = gql`
    query RepositoriesProgress {
        embeddingsSetupProgress {
            overallPercentComplete
            currentRepository
            currentRepositoryFilesProcessed
            currentRepositoryTotalFilesToProcess
            oneRepositoryReady
        }
    }
`

interface AppSetupProgressBarProps {
    /**
     * Whenever at least one repository has been processed
     * (embeddings uploading has been finished), Primary is used
     * for unblocking setup flow when we have at least one repo ready
     */
    onOneRepositoryFinished?: () => void
}

export const AppSetupProgressBar: FC<AppSetupProgressBarProps> = props => {
    const { onOneRepositoryFinished = noop } = props
    const { data } = useQuery<RepositoriesProgressResult>(REPO_UPLOADING_PROGRESS, {
        pollInterval: 2000,
        fetchPolicy: 'cache-and-network',
    })

    useEffect(() => {
        if (data && data.embeddingsSetupProgress.oneRepositoryReady) {
            onOneRepositoryFinished()
        }
    }, [data, onOneRepositoryFinished])

    if (
        !data ||
        data.embeddingsSetupProgress.overallPercentComplete === null ||
        data.embeddingsSetupProgress.overallPercentComplete === 100
    ) {
        return null
    }
    const currentRepository = data.embeddingsSetupProgress.currentRepository
    const filesProcessed = data.embeddingsSetupProgress.currentRepositoryFilesProcessed
    const filesToProcess = data.embeddingsSetupProgress.currentRepositoryTotalFilesToProcess

    const hasDetails = currentRepository && filesProcessed !== null && filesToProcess !== null

    return (
        <div className={styles.root}>
            <div className={styles.description}>
                {hasDetails && (
                    <>
                        <span>Generating embeddings for {data.embeddingsSetupProgress.currentRepository}</span>
                        <Text size="small" className={styles.percent}>
                            {data.embeddingsSetupProgress.currentRepositoryFilesProcessed} /{' '}
                            {data.embeddingsSetupProgress.currentRepositoryTotalFilesToProcess}
                        </Text>
                    </>
                )}

                {!hasDetails && (
                    <>
                        <span>Generating repositories embeddings</span>
                        <Text size="small" className={styles.percent}>
                            Loading files <LoadingSpinner />
                        </Text>
                    </>
                )}
            </div>

            <div className={styles.progress}>
                <div
                    className={styles.bar}
                    style={{ width: `${data.embeddingsSetupProgress.overallPercentComplete}%` }}
                />
            </div>
        </div>
    )
}
