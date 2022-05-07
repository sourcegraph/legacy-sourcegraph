import { RouteComponentProps } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { OrgAreaPageProps, OrgAreaRoute } from '../../org/area/OrgArea'
import { orgAreaRoutes } from '../../org/area/routes'
import { EditBatchSpecPageProps } from '../batches/batch-spec/edit/EditBatchSpecPage'
import { CreateOrEditBatchChangePageProps } from '../batches/create/CreateOrEditBatchChangePage'
import { NamespaceBatchChangesAreaProps, ExecutionAreaProps } from '../batches/global/GlobalBatchChangesArea'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'

const NamespaceBatchChangesArea = lazyComponent<NamespaceBatchChangesAreaProps, 'NamespaceBatchChangesArea'>(
    () => import('../batches/global/GlobalBatchChangesArea'),
    'NamespaceBatchChangesArea'
)

const ExecutionArea = lazyComponent<ExecutionAreaProps, 'ExecutionArea'>(
    () => import('../batches/global/GlobalBatchChangesArea'),
    'ExecutionArea'
)

const CreateOrEditBatchChangePage = lazyComponent<CreateOrEditBatchChangePageProps, 'CreateOrEditBatchChangePage'>(
    () => import('../batches/create/CreateOrEditBatchChangePage'),
    'CreateOrEditBatchChangePage'
)

const EditBatchSpecPage = lazyComponent<EditBatchSpecPageProps, 'EditBatchSpecPage'>(
    () => import('../batches/batch-spec/edit/EditBatchSpecPage'),
    'EditBatchSpecPage'
)

export const enterpriseOrganizationAreaRoutes: readonly OrgAreaRoute[] = [
    ...orgAreaRoutes,
    ...enterpriseNamespaceAreaRoutes,
    {
        path: '/batch-changes/:batchChangeName/edit',
        render: ({ match, ...props }: OrgAreaPageProps & RouteComponentProps<{ batchChangeName: string }>) => (
            <EditBatchSpecPage
                {...props}
                batchChange={{
                    name: match.params.batchChangeName,
                    url: match.url.replace('/edit', ''),
                    namespace: props.org,
                }}
            />
        ),
        condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
            batchChangesEnabled && batchChangesExecutionEnabled,
        fullPage: true,
    },
    {
        path: '/batch-changes/:batchChangeName/executions/:batchSpecID/configuration',
        render: ({ match, ...props }: OrgAreaPageProps & RouteComponentProps<{ batchChangeName: string }>) => (
            <CreateOrEditBatchChangePage
                {...props}
                initialNamespaceID={props.org.id}
                batchChangeName={match.params.batchChangeName}
                isReadOnly={true}
            />
        ),
        condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
            batchChangesEnabled && batchChangesExecutionEnabled,
        fullPage: true,
    },
    {
        path: '/batch-changes/:batchChangeName/executions/:batchSpecID',
        render: (props: OrgAreaPageProps & RouteComponentProps<{ batchSpecID: string }>) => (
            <ExecutionArea {...props} namespaceID={props.org.id} />
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
