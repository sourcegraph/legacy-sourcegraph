import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import {
    ExecuteBatchSpecFields,
    ExecuteBatchSpecResult,
    ExecuteBatchSpecVariables,
    BatchSpecByID2Result,
    BatchSpecByID2Variables,
    BatchSpecWorkspacesFields,
    CreateBatchSpecFromRawResult,
    CreateBatchSpecFromRawVariables,
    ReplaceBatchSpecInputResult,
    ReplaceBatchSpecInputVariables,
    Scalars,
} from '../../../graphql-operations'

export async function executeBatchSpec(spec: Scalars['ID']): Promise<ExecuteBatchSpecFields> {
    const result = await requestGraphQL<ExecuteBatchSpecResult, ExecuteBatchSpecVariables>(
        gql`
            mutation ExecuteBatchSpec($id: ID!) {
                executeBatchSpec(batchSpec: $id) {
                    ...ExecuteBatchSpecFields
                }
            }

            fragment ExecuteBatchSpecFields on BatchSpec {
                id
                namespace {
                    url
                }
            }
        `,
        { id: spec }
    ).toPromise()
    return dataOrThrowErrors(result).executeBatchSpec
}

const fragment = gql`
    fragment BatchSpecWorkspacesFields on BatchSpec {
        id
        originalInput
        workspaceResolution {
            workspaces {
                nodes {
                    ...BatchSpecWorkspaceFields
                }
            }
            state
            failureMessage
        }
        allowUnsupported
        allowIgnored
        importingChangesets {
            totalCount
            nodes {
                __typename
                id
                ... on VisibleChangesetSpec {
                    description {
                        __typename
                        ... on ExistingChangesetReference {
                            baseRepository {
                                name
                                url
                            }
                            externalID
                        }
                    }
                }
            }
        }
    }

    fragment BatchSpecWorkspaceFields on BatchSpecWorkspace {
        repository {
            id
            name
            url
        }
        ignored
        unsupported
        branch {
            id
            abbrevName
            displayName
            target {
                oid
            }
        }
        path
        onlyFetchWorkspace
        steps {
            run
            container
        }
        searchResultPaths
    }
`

export function createBatchSpecFromRaw(spec: string, namespace: Scalars['ID']): Observable<BatchSpecWorkspacesFields> {
    return requestGraphQL<CreateBatchSpecFromRawResult, CreateBatchSpecFromRawVariables>(
        gql`
            mutation CreateBatchSpecFromRaw($spec: String!, $namespace: ID!) {
                createBatchSpecFromRaw(batchSpec: $spec, namespace: $namespace) {
                    ...BatchSpecWorkspacesFields
                }
            }

            ${fragment}
        `,
        { spec, namespace }
    ).pipe(
        map(dataOrThrowErrors),
        map(result => result.createBatchSpecFromRaw)
    )
}

export function replaceBatchSpecInput(
    previousSpec: Scalars['ID'],
    spec: string
): Observable<BatchSpecWorkspacesFields> {
    return requestGraphQL<ReplaceBatchSpecInputResult, ReplaceBatchSpecInputVariables>(
        gql`
            mutation ReplaceBatchSpecInput($previousSpec: ID!, $spec: String!) {
                replaceBatchSpecInput(previousSpec: $previousSpec, batchSpec: $spec) {
                    ...BatchSpecWorkspacesFields
                }
            }

            ${fragment}
        `,
        { previousSpec, spec }
    ).pipe(
        map(dataOrThrowErrors),
        map(result => result.replaceBatchSpecInput)
    )
}

export function fetchBatchSpec(id: Scalars['ID']): Observable<BatchSpecWorkspacesFields> {
    return requestGraphQL<BatchSpecByID2Result, BatchSpecByID2Variables>(
        gql`
            query BatchSpecByID2($id: ID!) {
                node(id: $id) {
                    __typename
                    ...BatchSpecWorkspacesFields
                }
            }

            ${fragment}
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.node) {
                throw new Error('Not found')
            }
            if (data.node.__typename !== 'BatchSpec') {
                throw new Error(`Node is a ${data.node.__typename}, not a BatchSpec`)
            }
            return data.node
        })
    )
}
