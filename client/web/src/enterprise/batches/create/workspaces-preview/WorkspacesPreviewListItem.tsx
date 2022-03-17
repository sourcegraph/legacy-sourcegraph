import React, { useCallback, useMemo, useState } from 'react'

import { PreviewBatchSpecWorkspaceFields } from '../../../../graphql-operations'
import { CachedIcon, Descriptor, ExcludeButton, ExcludeIcon, ListItem } from '../../workspaces-list'

import styles from './WorkspacesPreviewListItem.module.scss'

interface WorkspacesPreviewListItemProps {
    workspace: PreviewBatchSpecWorkspaceFields
    /** Whether or not this item is stale */
    isStale: boolean
    /** Function to automatically update batch spec to exclude this item. */
    exclude: (repo: string, branch: string) => void
}

export const WorkspacesPreviewListItem: React.FunctionComponent<WorkspacesPreviewListItemProps> = ({
    workspace,
    isStale,
    exclude,
}) => {
    const [toBeExcluded, setToBeExcluded] = useState(false)

    const handleExclude = useCallback(() => {
        setToBeExcluded(true)
        exclude(workspace.repository.name, workspace.branch.displayName)
    }, [exclude, workspace])

    const statusIndicator = useMemo(
        () => (toBeExcluded ? <ExcludeIcon /> : workspace.cachedResultFound ? <CachedIcon /> : undefined),
        [toBeExcluded, workspace.cachedResultFound]
    )

    return (
        <ListItem className={isStale ? styles.stale : undefined}>
            <Descriptor workspace={workspace} statusIndicator={statusIndicator} />
            {toBeExcluded ? null : <ExcludeButton handleExclude={handleExclude} />}
        </ListItem>
    )
}
