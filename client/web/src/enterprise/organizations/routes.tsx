import React from 'react'
import { RouteComponentProps } from 'react-router'

import { OrgAreaPageProps, OrgAreaRoute } from '../../org/area/OrgArea'
import { orgAreaRoutes } from '../../org/area/routes'
import { lazyComponent } from '../../util/lazyComponent'
import { CreateBatchChangePage } from '../batches/create/CreateBatchChangePage'
import { NewCreateBatchChangePage } from '../batches/create/NewCreateBatchChangePage'
import { NamespaceBatchChangesAreaProps } from '../batches/global/GlobalBatchChangesArea'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'

const NamespaceBatchChangesArea = lazyComponent<NamespaceBatchChangesAreaProps, 'NamespaceBatchChangesArea'>(
    () => import('../batches/global/GlobalBatchChangesArea'),
    'NamespaceBatchChangesArea'
)

export const enterpriseOrganizationAreaRoutes: readonly OrgAreaRoute[] = [
    ...orgAreaRoutes,
    ...enterpriseNamespaceAreaRoutes,
    {
        path: '/batch-changes/create',
        render: props => <CreateBatchChangePage headingElement="h1" {...props} namespace={props.org} />,
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        fullPage: true,
    },
    {
        path: '/batch-changes/:batchChangeName/edit',
        render: ({ match, ...props }: OrgAreaPageProps & RouteComponentProps<{ batchChangeName: string }>) => (
            <NewCreateBatchChangePage {...props} namespace={props.org} batchChangeName={match.params.batchChangeName} />
        ),
        condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
            batchChangesEnabled && batchChangesExecutionEnabled,
        fullPage: true,
    },
    {
        path: '/batch-changes',
        render: props => <NamespaceBatchChangesArea {...props} namespaceID={props.org.id} />,
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
    },
]
