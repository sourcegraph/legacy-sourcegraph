import * as React from 'react'
import { useEffect } from 'react'

import classNames from 'classnames'

import { logger } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ErrorAlert, LoadingSpinner } from '@sourcegraph/wildcard'

import { FetchTreeOwnershipResult, FetchTreeOwnershipVariables } from '../../../graphql-operations'

import { FETCH_TREE_OWNERS } from './grapqlQueries'
import { OwnerList } from './OwnerList'

import styles from './FileOwnershipPanel.module.scss'
import {OwnershipAssignPermission} from '../../../rbac/constants';
import {MakeOwnerButton} from './MakeOwnerButton';
import {useFeatureFlag} from '../../../featureFlags/useFeatureFlag';

export interface OwnershipPanelProps {
    repoID: string
    revision?: string
    filePath: string
}

export const TreeOwnershipPanel: React.FunctionComponent<OwnershipPanelProps & TelemetryProps> = ({
    repoID,
    revision,
    filePath,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.log('OwnershipPanelOpened')
    }, [telemetryService])

    const { data, loading, error, refetch } = useQuery<FetchTreeOwnershipResult, FetchTreeOwnershipVariables>(
        FETCH_TREE_OWNERS,
        {
            variables: {
                repo: repoID,
                revision: revision ?? '',
                currentPath: filePath,
            },
        }
    )
    const [ownPromotionEnabled] = useFeatureFlag('own-promote')

    if (loading) {
        return (
            <div className={classNames(styles.loaderWrapper, 'text-muted')}>
                <LoadingSpinner inline={true} className="mr-1" /> Loading...
            </div>
        )
    }
    const canAssignOwners = (data?.currentUser?.permissions?.nodes || []).some(
        permission => permission.displayName === OwnershipAssignPermission
    )
    const makeOwnerButton =
        canAssignOwners && ownPromotionEnabled
            ? (userId: string | undefined) => (
                <MakeOwnerButton
                    onSuccess={refetch}
                    onError={() => {}} // TODO(#52911)
                    repoId={repoID}
                    path={filePath}
                    userId={userId}
                />
            )
            : undefined

    if (error) {
        logger.log(error)
        return (
            <div className={styles.contents}>
                <ErrorAlert error={error} prefix="Error getting ownership data" className="mt-2" />
            </div>
        )
    }

    if (data?.node?.__typename === 'Repository') {
        return <OwnerList data={data?.node?.commit?.tree?.ownership} isDirectory={true} makeOwnerButton={makeOwnerButton} />
    }
    return <OwnerList />
}
