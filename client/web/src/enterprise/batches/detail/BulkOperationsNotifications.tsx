import * as H from 'history'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { BulkOperationState } from '@sourcegraph/shared/src/graphql-operations'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { DismissibleAlert, isAlertDismissed } from '../../../components/DismissibleAlert'
import { ActiveBulkOperationsConnectionFields } from '../../../graphql-operations'
import { BatchChangeTab } from './BatchChangeTabs'

export interface BulkOperationsNotificationsProps {
    location: H.Location
    bulkOperations: ActiveBulkOperationsConnectionFields
}

export const BulkOperationsNotifications: React.FunctionComponent<BulkOperationsNotificationsProps> = ({
    bulkOperations,
    location,
}) => {
    // Don't show the header banners if the bulkoperations tab is open.
    const parameters = new URLSearchParams(location.search)
    if (parameters.get('tab') === BatchChangeTab.BULK_OPERATIONS) {
        return null
    }

    const latestProcessingNode = bulkOperations.nodes.find(node => node.state === BulkOperationState.PROCESSING)
    if (latestProcessingNode && !isAlertDismissed(`bulkOperation-processing-${latestProcessingNode.id}`)) {
        const processingCount = bulkOperations.nodes.filter(node => node.state === BulkOperationState.PROCESSING).length
        return (
            <DismissibleAlert
                className="alert alert-info"
                partialStorageKey={`bulkOperation-processing-${latestProcessingNode.id}`}
            >
                <span>
                    {processingCount} bulk {pluralize('operation', processingCount)}{' '}
                    {pluralize('is', processingCount, 'are')} currently running. Click the{' '}
                    <Link to="?tab=bulkoperations">bulk operations tab</Link> to view.
                </span>
            </DismissibleAlert>
        )
    }

    const latestFailedNode = bulkOperations.nodes.find(node => node.state === BulkOperationState.FAILED)
    if (latestFailedNode && !isAlertDismissed(`bulkOperation-failed-${latestFailedNode.id}`)) {
        const failedCount = bulkOperations.nodes.filter(node => node.state === BulkOperationState.FAILED).length
        return (
            <DismissibleAlert
                className="alert alert-info"
                partialStorageKey={`bulkOperation-failed-${latestFailedNode.id}`}
            >
                <span>
                    {failedCount} bulk {pluralize('operation', failedCount)} {pluralize('has', failedCount, 'have')}{' '}
                    recently failed running. Click the <Link to="?tab=bulkoperations">bulk operations tab</Link> to
                    view.
                </span>
            </DismissibleAlert>
        )
    }
    const latestCompleteNode = bulkOperations.nodes.find(node => node.state === BulkOperationState.COMPLETED)
    if (latestCompleteNode && !isAlertDismissed(`bulkOperation-completed-${latestCompleteNode.id}`)) {
        const completeCount = bulkOperations.nodes.filter(node => node.state === BulkOperationState.COMPLETED).length
        return (
            <DismissibleAlert
                className="alert alert-info"
                partialStorageKey={`bulkOperation-completed-${latestCompleteNode.id}`}
            >
                <span>
                    {completeCount} bulk {pluralize('operation', completeCount)}{' '}
                    {pluralize('has', completeCount, 'have')} recently finished running. Click the{' '}
                    <Link to="?tab=bulkoperations">bulk operations tab</Link> to view.
                </span>
            </DismissibleAlert>
        )
    }
    return null
}
