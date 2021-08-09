import * as H from 'history'
import BookOpenVariantIcon from 'mdi-react/BookOpenVariantIcon'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import LinkVariantIcon from 'mdi-react/LinkVariantIcon'
import React, { RefObject, useEffect, useMemo, useRef } from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'

import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { AnchorLink } from '@sourcegraph/shared/src/components/Link'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { ResolvedRevisionSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { Badge } from '../../components/Badge'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { RepositoryFields } from '../../graphql-operations'
import { toDocumentationURL } from '../../util/url'

import { DocumentationExamples } from './DocumentationExamples'
import { DocumentationIcons } from './DocumentationIcons'
import { GQLDocumentationNode, Tag, isExcluded } from './graphql'

interface Props
    extends Partial<RevisionSpec>,
        ResolvedRevisionSpec,
        BreadcrumbSetters,
        SettingsCascadeProps,
        VersionContextProps {
    repo: RepositoryFields

    history: H.History
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    commitID: string

    /** The documentation node to render */
    node: GQLDocumentationNode

    /** How far deep we are in the tree of documentation nodes */
    depth: number

    /** Whether or not this is the first child of a parent node */
    isFirstChild: boolean

    /** The pathID of the page containing this documentation node */
    pagePathID: string

    /** A list of documentation tags, a section will not be rendered if it matches one of these. */
    excludingTags: Tag[]

    /** Called to get the current scrollTop position of the scrolling page */
    currentScrollTop: () => number

    /**
     * Called when this documentation node becomes visible / scrolled into the screen, or becomes
     * invisible.
     *
     * The provided number indicates the position of the node on the page, with lesser numbers
     * being at the top and greater numbers being at the bottom.
     **/
    onVisibilityChange: (visible: boolean, node: GQLDocumentationNode, top: number) => void
}

export const DocumentationNode: React.FunctionComponent<Props> = React.memo(({
    useBreadcrumb,
    node,
    depth,
    isFirstChild,
    currentScrollTop,
    onVisibilityChange,
    ...props
}) => {
    const repoRevision = {
        repoName: props.repo.name,
        revision: props.revision || '',
    }
    const hashIndex = node.pathID.indexOf('#')
    const hash = hashIndex !== -1 ? node.pathID.slice(hashIndex + '#'.length) : ''
    let path = hashIndex !== -1 ? node.pathID.slice(0, hashIndex) : node.pathID
    path = path === '/' ? '' : path
    const thisPage = toDocumentationURL({ ...repoRevision, pathID: path + '#' + hash })

    useBreadcrumb(
        useMemo(
            () =>
                depth === 0 ? { key: `node-${depth}`, element: <Link to={thisPage}>{node.label.value}</Link> } : null,
            [depth, node.label.value, thisPage]
        )
    )
    if (node.detail.value === '') {
        const children = node.children.filter(child =>
            !child.node ? false : !isExcluded(child.node, props.excludingTags)
        )
        if (children.length === 0) {
            return null
        }
    }

    const headingLevel = depth + 1 < 4 ? depth + 1 : 4
    const topMargin =
        depth === 0
            ? ' mt-3' // Level 0 header ("Package foo")
            : depth === 1
            ? ' mt-5' // Level 1 headers ("Constants", "Variables", etc.)
            : isFirstChild
            ? ' mt-4'
            : ' mt-5' // Lowest level headers

    const ref: RefObject<HTMLHeadingElement|undefined>|null|undefined = useRef();
    const intersectionObserver = new IntersectionObserver(
        ([entry]) => {
            if (entry.isIntersecting) {
                onVisibilityChange(true, node, currentScrollTop() + entry.boundingClientRect.top)
            } else {
                onVisibilityChange(false, node, 0)
            }
        }
    )
    useEffect(() => {
        if (ref.current) intersectionObserver.observe(ref.current)
        // Remove the observer as soon as the component is unmounted
        return () => { intersectionObserver.disconnect() }
    }, [])

    return (
        <div className={`documentation-node mb-5${topMargin}`}>
            <Heading level={headingLevel} className="d-flex align-items-center documentation-node__heading" innerRef={ref}>
                <AnchorLink className="documentation-node__heading-anchor-link" to={thisPage}>
                    <LinkVariantIcon className="icon-inline" />
                </AnchorLink>
                {depth !== 0 && <DocumentationIcons className="mr-1" tags={node.documentation.tags} />}
                <Link className="h" id={hash} to={thisPage}>
                    {node.label.value}
                </Link>
            </Heading>
            {depth === 0 && (
                <>
                    <div className="d-flex align-items-center mb-3">
                        <span className="documentation-node__pill d-flex justify-content-center align-items-center px-2">
                            <BookOpenVariantIcon className="icon-inline text-muted mr-1" /> Generated API docs
                            <span className="documentation-node__pill-divider mx-2" />
                            <a
                                // eslint-disable-next-line react/jsx-no-target-blank
                                target="_blank"
                                rel="noopener"
                                href="https://docs.sourcegraph.com/code_intelligence/apidocs"
                            >
                                Learn more
                            </a>
                        </span>
                        {/*
                        TODO(apidocs): add support for indicating time the API docs were updated
                        <span className="ml-2">Last updated 2 days ago</span>
                    */}
                        <Badge status="experimental" className="text-uppercase ml-2" />
                    </div>
                    <hr />
                </>
            )}
            {node.detail.value !== '' && (
                <div className="pt-2">
                    <Markdown dangerousInnerHTML={renderMarkdown(node.detail.value)} />
                </div>
            )}

            {!isExcluded(node, ['test', 'benchmark', 'example', 'license', 'owner', 'package']) &&
                node.documentation.tags.length !== 0 && (
                    <>
                        <h4 className="mt-4">
                            Usage examples
                            <HelpCircleOutlineIcon
                                className="icon-inline ml-1"
                                data-tooltip="Usage examples from precise LSIF code intelligence index"
                            />
                        </h4>
                        <DocumentationExamples {...props} pathID={node.pathID} />
                    </>
                )}

            {node.children?.map(
                (child, index) =>
                    child.node &&
                    !isExcluded(child.node, props.excludingTags) && (
                        <DocumentationNode
                            key={`${depth}-${child.node.pathID}`}
                            {...props}
                            node={child.node}
                            depth={depth + 1}
                            isFirstChild={index === 0}
                            useBreadcrumb={useBreadcrumb}
                            currentScrollTop={currentScrollTop}
                            onVisibilityChange={onVisibilityChange}
                        />
                    )
            )}
        </div>
    )
})

const Heading: React.FunctionComponent<{
    level: number
    children: React.ReactNode
    innerRef: RefObject<any>
    [x:string]: any;
}> = ({
    level,
    children,
    innerRef,
    ...props
}) => {
    return React.createElement(
        `h${level}`,
        {ref: innerRef, ...props},
        children,
    )
}
