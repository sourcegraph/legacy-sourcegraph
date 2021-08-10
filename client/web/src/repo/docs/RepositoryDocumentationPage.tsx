import * as H from 'history'
import { upperFirst } from 'lodash'
import BookOpenVariantIcon from 'mdi-react/BookOpenVariantIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useEffect, useCallback, useMemo, useState, useRef, RefObject } from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { asError, ErrorLike } from '@sourcegraph/shared/src/util/errors'
import { RevisionSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container, useDebounce } from '@sourcegraph/wildcard'

import { Badge } from '../../components/Badge'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { PageTitle } from '../../components/PageTitle'
import { useScrollToLocationHash } from '../../components/useScrollToLocationHash'
import { RepositoryFields } from '../../graphql-operations'
import { FeedbackPrompt } from '../../nav/Feedback/FeedbackPrompt'
import { routes } from '../../routes'
import { eventLogger } from '../../tracking/eventLogger'
import { toDocumentationURL } from '../../util/url'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'

import { DocumentationNode } from './DocumentationNode'
import { DocumentationWelcomeAlert } from './DocumentationWelcomeAlert'
import { fetchDocumentationPage, fetchDocumentationPathInfo, GQLDocumentationNode, isExcluded, Tag } from './graphql'
import { RepositoryDocumentationSidebar, getSidebarVisibility } from './RepositoryDocumentationSidebar'

const PageError: React.FunctionComponent<{ error: ErrorLike }> = ({ error }) => (
    <div className="repository-docs-page__error alert alert-danger m-2">Error: {upperFirst(error.message)}</div>
)

const PageNotFound: React.FunctionComponent = () => (
    <div className="repository-docs-page__not-found">
        <MapSearchIcon className="icon-inline" /> Page not found
    </div>
)

interface Props
    extends RepoHeaderContributionsLifecycleProps,
        Partial<RevisionSpec>,
        ResolvedRevisionSpec,
        BreadcrumbSetters,
        SettingsCascadeProps,
        VersionContextProps {
    repo: RepositoryFields
    history: H.History
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    pathID: string
    commitID: string
}

const LOADING = 'loading' as const

/** A page that shows a repository's documentation at the current revision. */
export const RepositoryDocumentationPage: React.FunctionComponent<Props> = React.memo(function Render({
    useBreadcrumb,
    ...props
}) {
    useEffect(() => {
        eventLogger.logViewEvent('RepositoryDocs')
    }, [])
    useScrollToLocationHash(props.location)

    const thisPage = toDocumentationURL({ repoName: props.repo.name, revision: props.revision || '', pathID: '' })
    useBreadcrumb(useMemo(() => ({ key: 'node', element: <Link to={thisPage}>API docs</Link> }), [thisPage]))

    const pagePathID = props.pathID || '/'
    const page =
        useObservable(
            useMemo(
                () =>
                    fetchDocumentationPage({
                        repo: props.repo.id,
                        revspec: props.commitID,
                        pathID: pagePathID,
                    }).pipe(
                        catchError(error => [asError(error)]),
                        startWith(LOADING)
                    ),
                [props.repo.id, props.commitID, pagePathID]
            )
        ) || LOADING

    const pathInfo =
        useObservable(
            useMemo(
                () =>
                    fetchDocumentationPathInfo({
                        repo: props.repo.id,
                        revspec: props.commitID,
                        pathID: pagePathID,
                        ignoreIndex: true,
                        maxDepth: 1,
                    }).pipe(
                        catchError(error => [asError(error)]),
                        startWith(LOADING)
                    ),
                [props.repo.id, props.commitID, pagePathID]
            )
        ) || LOADING

    const [sidebarVisible, setSidebarVisible] = useState(getSidebarVisibility())
    const handleSidebarVisible = useCallback((visible: boolean) => setSidebarVisible(visible), [])

    const loading = page === LOADING || pathInfo === LOADING
    const error = isErrorLike(page) ? page : isErrorLike(pathInfo) ? pathInfo : null

    const excludingTags: Tag[] = useMemo(() => ['private'], [])

    const containerReference: RefObject<HTMLDivElement | undefined> | null | undefined = useRef()

    // Keep track of all node headers on the page that are visible. When visibility changes, sort
    // them based on position so we can determine the top-most header that is visible on the page
    // and update the active node - which is used for various visual effects.
    const [activePathID, setActivePathID] = useState<string | null>(null)

    const [_, setVisiblePathIDs] = useState<{ top: number; pathID: string }[]>([]);
    const onVisibilityChange = React.useMemo(() =>
        (visible: boolean, node: GQLDocumentationNode, top: number): void => {
            setVisiblePathIDs(visiblePathIDs => {
                let updated = visiblePathIDs;
                if (!visible) {
                    updated = updated.filter(pair => pair.pathID !== node.pathID)
                } else {
                    updated.push({ top, pathID: node.pathID })
                    updated.sort((a, b) => (a.top < b.top ? -1 : 1))
                }
                if (updated.length > 0) {
                    setActivePathID(() => updated[0].pathID)
                }
                return updated;
            })
        },
        [setVisiblePathIDs, setActivePathID],
    )

    const currentScrollTop = React.useMemo(() =>
        () => containerReference?.current?.scrollTop || 0,
        [containerReference],
    )
    return (
        <div className="repository-docs-page">
            <PageTitle title="API docs" />
            {loading ? <LoadingSpinner className="icon-inline m-1" /> : null}
            {error && error.message === 'page not found' ? <PageNotFound /> : null}
            {error && (error.message === 'no LSIF data' || error.message === 'no LSIF documentation') ? (
                <div className="repository-docs-page__container">
                    <div className="repository-docs-page__container-content">
                        <div className="d-flex float-right">
                            <a
                                // eslint-disable-next-line react/jsx-no-target-blank
                                target="_blank"
                                rel="noopener"
                                href="https://docs.sourcegraph.com/code_intelligence/apidocs"
                                className="mr-1 btn btn-sm text-decoration-none btn-link btn-outline-secondary"
                            >
                                Learn more
                            </a>
                            <FeedbackPrompt routes={routes} />
                        </div>
                        <h1>
                            <BookOpenVariantIcon className="icon-inline mr-1" />
                            API docs
                            <Badge status="experimental" className="text-uppercase ml-2" />
                        </h1>
                        <p>API documentation generated for all your code</p>
                        <Container>
                            <h2 className="text-muted mb-2">
                                <MapSearchIcon className="icon-inline mr-2" />
                                Repository has no LSIF documentation data
                            </h2>
                            <p className="mt-3">
                                Sourcegraph can use LSIF code intelligence to generate API documentation for all your
                                code, giving you the ability to navigate and explore the APIs provided by this
                                repository.
                            </p>
                            <h3>
                                <a
                                    // eslint-disable-next-line react/jsx-no-target-blank
                                    target="_blank"
                                    rel="noopener"
                                    href="https://docs.sourcegraph.com/code_intelligence/apidocs"
                                >
                                    Learn more
                                </a>
                            </h3>
                            <p className="text-muted mt-3 mb-0">
                                <strong>Note:</strong> only the Go programming language is currently supported.
                            </p>
                        </Container>
                    </div>
                </div>
            ) : null}
            {isErrorLike(error) &&
            error.message !== 'page not found' &&
            error.message !== 'no LSIF data' &&
            error.message !== 'no LSIF documentation' ? (
                <PageError error={error} />
            ) : null}
            {page !== LOADING && !isErrorLike(page) && pathInfo !== LOADING && !isErrorLike(pathInfo) ? (
                <>
                    <RepositoryDocumentationSidebar
                        {...props}
                        onToggle={handleSidebarVisible}
                        node={page.tree}
                        pathInfo={pathInfo}
                        pagePathID={pagePathID}
                        activePathID={activePathID || pagePathID}
                        depth={0}
                    />
                    <div className="repository-docs-page__container" ref={containerReference}>
                        <div
                            className={`repository-docs-page__container-content${
                                sidebarVisible ? ' repository-docs-page__container-content--sidebar-visible' : ''
                            }`}
                        >
                            {/*
                                TODO(apidocs): Eventually this welcome alert should go away entirely, but for now
                                it's the best thing we have for the sometimes empty root landing page.
                            */}
                            {page.tree.detail.value === '' && <DocumentationWelcomeAlert />}
                            {isExcluded(page.tree, excludingTags) ? (
                                <div className="m-3">
                                    <h2 className="text-muted">Looks like there's nothing to see here.</h2>
                                    <p>API docs for private / unexported code is coming soon!</p>
                                </div>
                            ) : null}
                            <DocumentationNode
                                {...props}
                                useBreadcrumb={useBreadcrumb}
                                node={page.tree}
                                pagePathID={pagePathID}
                                depth={0}
                                isFirstChild={true}
                                excludingTags={excludingTags}
                                currentScrollTop={currentScrollTop}
                                onVisibilityChange={onVisibilityChange}
                            />
                        </div>
                    </div>
                </>
            ) : null}
        </div>
    )
})
