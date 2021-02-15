import React, { useEffect } from 'react'
import { Observable } from 'rxjs'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../backend/graphql'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { map } from 'rxjs/operators'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { RouteComponentProps } from 'react-router'
import { SymbolsSidebarOptionsSetterProps } from './SymbolsArea'
import {
    DocSymbolFieldsFragment,
    DocumentSymbolResult,
    DocumentSymbolVariables,
    SymbolPageSymbolFields,
} from '../../graphql-operations'
import { Markdown } from '../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../shared/src/util/markdown'

export interface Symbol extends DocSymbolFieldsFragment {
    children?: Symbol[]
}

const SymbolPageSymbolsGQLFragment = gql`
    fragment DocSymbolFieldsFragment on DocSymbol {
        id
        text
        detail
        kind
        tags
        hover {
            markdown {
                text
            }
        }
    }
    fragment DocSymbolHierarchyFragment on DocSymbol {
        ...DocSymbolFieldsFragment
        children {
            ...DocSymbolFieldsFragment
            children {
                ...DocSymbolFieldsFragment
                children {
                    ...DocSymbolFieldsFragment
                }
            }
        }
    }
    fragment SymbolPageSymbolFields on DocSymbol {
        ...DocSymbolHierarchyFragment
        root {
            ...DocSymbolHierarchyFragment
        }
    }
`

const querySymbolUncached = (vars: DocumentSymbolVariables): Observable<SymbolPageSymbolFields | null | undefined> =>
    requestGraphQL<DocumentSymbolResult, DocumentSymbolVariables>(
        gql`
            query DocumentSymbol($repo: ID!, $commitID: String!, $symbolID: String!) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $commitID) {
                            tree(path: "") {
                                docSymbol(id: $symbolID) {
                                    ...SymbolPageSymbolFields
                                }
                            }
                        }
                    }
                }
            }
            ${SymbolPageSymbolsGQLFragment}
        `,
        vars
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node?.commit?.tree?.docSymbol)
    )

const querySymbol = memoizeObservable(querySymbolUncached, parameters => JSON.stringify(parameters))

export interface SymbolRouteProps {
    symbolID: string
}

export interface Props
    extends Pick<RepoRevisionContainerContext, 'repo' | 'revision'>,
        SymbolsSidebarOptionsSetterProps,
        RouteComponentProps<SymbolRouteProps> {}

export const SymbolPage: React.FunctionComponent<Props> = ({
    repo,
    revision,
    match: {
        params: { symbolID },
    },
    history,
    setSidebarOptions,
}) => {
    const symbol = useObservable(querySymbol({ repo: repo.id, commitID: revision, symbolID }))
    useEffect(() => {
        setSidebarOptions({ containerSymbol: symbol?.root as Symbol })
        return () => setSidebarOptions(null)
    }, [symbol || null]) // TODO(beyang): may want to specify dependencies
    if (!symbol) {
        return <div>Symbol not found</div>
    }

    const hoverParts = symbol.hover?.markdown.text.split('---', 2)
    const hoverSig = hoverParts?.[0]
    const hoverDoc = hoverParts?.[1]

    return (
        <>
            {hoverSig && (
                <Markdown
                    dangerousInnerHTML={renderMarkdown(hoverSig)}
                    history={history}
                    className={`symbol-hover__signature`}
                />
            )}
            {hoverDoc && <Markdown dangerousInnerHTML={renderMarkdown(hoverDoc)} history={history} />}
            <div>Symbol: {symbol.text}</div>
            <div>Definition</div>
            <div>Detail: {symbol.detail}</div>
            <div>Examples</div>
            <div>Children</div>
        </>
    )
}
