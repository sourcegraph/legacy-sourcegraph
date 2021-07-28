import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import classNames from 'classnames'
import { escapeRegExp } from 'lodash'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import React, { ReactElement, useCallback, useMemo, useState } from 'react'
import { Collapse } from 'reactstrap'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { FILTERS, FilterType, isNegatableFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../..'
import { QueryChangeSource, QueryState } from '../../helpers'
import {
    createFilterExampleFromString,
    updateQueryWithFilterAndExample,
    QueryExample,
} from '../../helpers/examplevalue'

import styles from './SearchReference.module.scss'
import sidebarStyles from './SearchSidebarSection.module.scss'

const SEARCH_REFERENCE_TAB_KEY = 'SearchProduct.SearchReference.Tab'

type FilterInfo = QueryExample & {
    field: FilterType
    description: string
    /**
     * Force showing or not showing suggestions for this fileter.
     */
    showSuggestions?: boolean
    /**
     * Used to indicate whether this filter/example should be listed in the
     * "Common" filters section and at which position
     */
    commonRank?: number
    alias?: string
    examples?: string[]
}

type  OperatorInfo  = QueryExample & {
    operator: string
    description: string
    alias?: string
    examples?: string[]
}

type SearchReferenceInfo = FilterInfo | OperatorInfo

/**
 * Adds additional search reference information from the existing filters list.
 */
function augmentFilterInfo(searchReference: FilterInfo): void {
    const filter = FILTERS[searchReference.field]
    if (filter?.alias) {
        searchReference.alias = filter.alias
    }
}

const filterInfos: FilterInfo[] = [
    {
        ...createFilterExampleFromString('"{last week}"'),
        field: FilterType.after,
        description:
            'Only include results from diffs or commits which have a commit date after the specified time frame. To use this filter, the search query must contain `type:diff` or `type:commit`.',
        commonRank: 100,
        examples: ['after:"6 weeks ago"', 'after:"november 1 2019"'],
    },
    {
        ...createFilterExampleFromString('{yes/only}'),
        field: FilterType.archived,
        description:
            'The "yes" option includes archived repositories. The "only" option filters results to only archived repositories. Results in archived repositories are excluded by default.',
        examples: ['repo:sourcegraph/ archived:only'],
    },
    {
        ...createFilterExampleFromString('{name}'),
        field: FilterType.author,
        description: `Only include results from diffs or commits authored by the user. Regexps are supported. Note that they match the whole author string of the form \`Full Name <user@example.com>\`, so to include only authors from a specific domain, use \`author:example.com>$\`.

You can also search by \`committer:git-email\`. *Note: there is a committer only when they are a different user than the author.*

To use this filter, the search query must contain \`type:diff\` or \`type:commit\`.`,
        examples: ['type:diff author:nick'],
    },
    {
        ...createFilterExampleFromString('"{last thursday}"'),
        field: FilterType.before,
        description:
            'Only include results from diffs or commits which have a commit date before the specified time frame. To use this filter, the search query must contain `type:diff` or `type:commit`.',
        commonRank: 100,
        examples: ['before:"last thursday"', 'before:"november 1 2019"'],
    },
    {
        ...createFilterExampleFromString('{yes}'),
        field: FilterType.case,
        description: 'Perform a case sensitive query. Without this, everything is matched case insensitively.',
        examples: ['OPEN_FILE case:yes'],
    },
    {
        ...createFilterExampleFromString('"{pattern}"'),
        field: FilterType.content,
        description:
            'Set the search pattern with a dedicated parameter. Useful when searching literally for a string that may conflict with the search pattern syntax. In between the quotes, the `\\` character will need to be escaped (`\\\\` to evaluate for `\\`).',
        commonRank: 70,
        examples: ['repo:sourcegraph content:"repo:sourcegraph"', 'file:Dockerfile alpine -content:alpine:latest'],
    },
    {
        ...createFilterExampleFromString('{N/all}'),
        field: FilterType.count,
        description:
            'Retrieve *N* results. By default, Sourcegraph stops searching early and returns if it finds a full page of results. This is desirable for most interactive searches. To wait for all results, use **count:all**.',
        commonRank: 60,
        examples: ['count:1000 function', 'count:all err'],
    },
    {
        ...createFilterExampleFromString('{regexp-pattern}'),
        field: FilterType.file,
        commonRank: 30,
        description: 'Only include results in files whose full path matches the regexp.',
        examples: ['file:.js$ httptest', 'file:internal/ httptest', 'file:.js$ -file:test http'],
    },
    {
        ...createFilterExampleFromString('contains.content({regexp-pattern})'),
        field: FilterType.file,
        description: 'Search only inside files that contain content matching the provided regexp pattern.',
        examples: ['file:contains.content(github.com/sourcegraph/sourcegraph)'],
    },
    {
        ...createFilterExampleFromString('{yes/only}'),
        field: FilterType.fork,
        description:
            'Include results from repository forks or filter results to only repository forks. Results in repository forks are exluded by default.',
        commonRank: 80,
        examples: ['fork:yes repo:sourcegraph'],
    },
    {
        ...createFilterExampleFromString('{language-name}'),
        field: FilterType.lang,
        description: 'Only include results from files in the specified programming language.',
        commonRank: 40,
        examples: ['lang:typescript encoding', '-lang:typescript encoding'],
    },
    {
        ...createFilterExampleFromString('"{any string}"'),
        field: FilterType.message,
        description: `Only include results from diffs or commits which have commit messages containing the string.

To use this filter, the search query must contain \`type:diff\` or \`type:commit\`.`,
        examples: ['type:commit message:"testing"', 'type:diff message:"testing"'],
    },
    {
        ...createFilterExampleFromString('{regexp-pattern}'),
        field: FilterType.repo,
        description:
            'Only include results from repositories whose path matches the regexp-pattern. A repository’s path is a string such as *github.com/myteam/abc* or *code.example.com/xyz* that depends on your organization’s repository host. If the regexp ends in `@rev`, that revision is searched instead of the default branch (usually `master`). `repo:regexp-pattern@rev` is equivalent to `repo:regexp-pattern rev:rev`.',
        commonRank: 10,
        examples: [
            'repo:gorilla/mux testroute',
            'repo:^github.com/sourcegraph/sourcegraph$@v3.14.0 mux',
            'repo:alice/ -repo:old-repo',
            'repo:vscode@*refs/heads/:^refs/heads/master type:diff task',
        ],
    },
    {
        ...createFilterExampleFromString('{group-name}'),
        field: FilterType.repogroup,
        description:
            'Only include results from the named group of repositories (defined by the server admin). Same as using a repo: keyword that matches all of the group’s repositories. Use repo: unless you know that the group exists.',
    },
    {
        ...createFilterExampleFromString('contains.file({path})'),
        field: FilterType.repo,
        description: 'Search only inside repositories that contain a file path matching the regular expression.',
        examples: ['repo:contains.file(README)'],
        showSuggestions: false,
    },
    {
        ...createFilterExampleFromString('contains.content({content})'),
        field: FilterType.repo,
        description: 'Search only inside repositories that contain file content matching the regular expression.',
        examples: ['repo:contains.content(TODO)'],
        showSuggestions: false,
    },
    {
        ...createFilterExampleFromString('contains({file:path content:content})'),
        field: FilterType.repo,
        description:
            'Search only inside repositories that contain a file matching the `file:` with `content:` filters.',
        examples: ['repo:contains(file:CHANGELOG content:fix)'],
        showSuggestions: false,
    },
    {
        ...createFilterExampleFromString('contains.commit.after({date})'),
        field: FilterType.repo,
        description:
            'Search only inside repositories that contain a a commit after some specified time. See [git date formats](https://github.com/git/git/blob/master/Documentation/date-formats.txt) for accepted formats. Use this to filter out stale repositories that don’t contain commits past the specified time frame. This parameter is experimental.',
        examples: ['repo:contains.commit.after(1 month ago)', 'repo:contains.commit.after(june 25 2017)'],
        showSuggestions: false,
    },
    {
        ...createFilterExampleFromString('{revision}'),
        field: FilterType.rev,
        commonRank: 20,
        description:
            'Search a revision instead of the default branch. `rev:` can only be used in conjunction with `repo:` and may not be used more than once. See our [revision syntax documentation](https://docs.sourcegraph.com/code_search/reference/queries#repository-revisions) to learn more.',
    },
    {
        ...createFilterExampleFromString('{result-types}'),
        field: FilterType.select,
        commonRank: 50,
        description:
            'Shows only query results for a given type. For example, `select:repo` displays only distinct reopsitory paths from search results. See [language definition](https://docs.sourcegraph.com/code_search/reference/language#select) for possible values.',
        examples: ['fmt.Errorf select:repo'],
    },
    {
        ...createFilterExampleFromString('{diff/commit/...}'),
        field: FilterType.type,
        commonRank: 1,
        description:
            'Specifies the type of search. By default, searches are executed on all code at a given point in time (a branch or a commit). Specify the `type:` if you want to search over changes to code or commit messages instead (diffs or commits).',
        examples: ['type:symbol path', 'type:diff func', 'type:commit test'],
    },
    {
        ...createFilterExampleFromString('{golang-duration-value}'),
        field: FilterType.timeout,
        description:
            'Customizes the timeout for searches. The value of the parameter is a string that can be parsed by the [Go time package’s `ParseDuration`](https://golang.org/pkg/time/#ParseDuration) (e.g. 10s, 100ms). By default, the timeout is set to 10 seconds, and the search will optimize for returning results as soon as possible. The timeout value cannot be set longer than 1 minute. When provided, the search is given the full timeout to complete.',
        examples: ['repo:^github.com/sourcegraph timeout:15s func count:10000'],
    },
    {
        ...createFilterExampleFromString('{any/private/public}'),
        field: FilterType.visibility,
        description:
            'Filter results to only public or private repositories. The default is to include both private and public repositories.',
        examples: ['type:repo visibility:public'],
    },
]

for (const info of filterInfos) {
    augmentFilterInfo(info)
}

const commonFilters = filterInfos
    .filter(info => info.commonRank !== undefined)
    // commonRank will never be undefined here, but TS doesn't seem to know
    .sort((a, b) => (a.commonRank as number) - (b.commonRank as number))

const operatorInfo: OperatorInfo[] = [
    {
        ...createFilterExampleFromString('{expr} AND {expr}'),
        operator: 'AND',
        alias: 'and',
        description:
            'Returns results for files containing matches on the left and right side of the `and` (set intersection).',
        examples: ['conf.Get( and log15.Error(', 'conf.Get( AND log15.Error( AND after'],
    },
    {
        ...createFilterExampleFromString('({expr} OR {expr})'),
        operator: 'OR',
        alias: 'or',
        description:
            'Returns file content matching either on the left or right side, or both (set union). The number of results reports the number of matches of both strings.',
        examples: ['conf.Get( or log15.Error(', 'conf.Get( OR log15.Error( OR after'],
    },
    {
        ... createFilterExampleFromString('NOT {expr}'),
        operator: 'NOT',
        alias: 'not',
        description:
            '`NOT` can be prepended to negate filters like `file`, `lang`, `repo`. Prepending `NOT` to search patterns excludes documents that contain the pattern. For readability, you may use `NOT` in conjunction with `AND` if you like: `panic AND NOT ever`.',
        examples: ['lang:go not file:main.go panic', 'panic NOT ever'],
    },
]

/**
 * Returns true if the provided regular expressions all match the provided
 * filter information (name, description, ...)
 */
function matches(searchTerms: RegExp[], info: FilterInfo): boolean {
    return searchTerms.every(term => term.test(info.field) || term.test(info.description || ''))
}

/**
 * Convert the search input into an array of regular expressions. Each word in
 * the input becomes a regular expression starting with a word boundary check.
 */
function parseSearchInput(searchInput: string): RegExp[] {
    const terms = searchInput.split(/\s+/)
    return terms.map(term => new RegExp(`\\b${escapeRegExp(term)}`))
}

/**
 * Whether or not to trigger the suggestion popover when adding this filter to
 * the query.
 */
function shouldShowSuggestions(searchReference: FilterInfo): boolean {
    return Boolean(searchReference.showSuggestions !== false && FILTERS[searchReference.field].discreteValues)
}

function isFilterInfo(searchReference: SearchReferenceInfo): searchReference is FilterInfo {
    return (searchReference as FilterInfo).field !== undefined
}

const classNameTokenMap = {
    text: 'search-filter-keyword',
    placeholder: styles.placeholder,
}

interface SearchReferenceExampleProps {
    example: string
    onClick?: (example: string) => void
}

const SearchReferenceExample: React.FunctionComponent<SearchReferenceExampleProps> = ({ example, onClick }) => {
    // All current examples are literal queries
    const scanResult = scanSearchQuery(example, false, SearchPatternType.literal)
    // We only use valid queries as examples, so this will always be true
    if (scanResult.type === 'success') {
        return (
            <button className="btn p-0 flex-1" type="button" onClick={() => onClick?.(example)}>
                {scanResult.term.map((term, index) => {
                    switch (term.type) {
                        case 'filter':
                            return (
                                <React.Fragment key={index}>
                                    <span className="search-filter-keyword">{term.field.value}:</span>
                                    {term.value?.quoted ? `"${term.value.value}"` : term.value?.value}
                                </React.Fragment>
                            )
                        case 'keyword':
                            return (
                                <span key={index} className="search-filter-keyword">
                                    {term.value}
                                </span>
                            )
                        default:
                            return example.slice(term.range.start, term.range.end)
                    }
                })}
            </button>
        )
    }
    return null
}

interface SearchReferenceEntryProps<T extends SearchReferenceInfo> {
    searchReference: T
    onClick: (searchReference: T, negate: boolean) => void
    onExampleClick?: (example: string) => void
}

const SearchReferenceEntry = <T extends SearchReferenceInfo>({
    searchReference,
    onClick,
    onExampleClick,
}: SearchReferenceEntryProps<T>): ReactElement | null => {
    const [collapsed, setCollapsed] = useState(true)
    const CollapseIcon = collapsed ? ChevronLeftIcon : ChevronDownIcon

    let buttonTextPrefix: ReactElement | null = null
    if (isFilterInfo(searchReference)) {
        buttonTextPrefix = <span className="search-filter-keyword">{searchReference.field}:</span>
    }

    return (
        <li>
            <span
                className={classNames(styles.item, sidebarStyles.sidebarSectionListItem, {
                    [styles.active]: !collapsed,
                })}
            >
                <button
                    className="btn p-0 flex-1"
                    type="button"
                    onClick={event => onClick(searchReference, event.altKey)}
                >
                    <span className="text-monospace">
                        {buttonTextPrefix}
                        {searchReference.tokens.map(token => (
                            <span key={token.start} className={classNameTokenMap[token.type]}>
                                {token.value}
                            </span>
                        ))}
                    </span>
                </button>
                <button
                    type="button"
                    className={classNames('btn btn-icon', styles.collapseButton)}
                    onClick={event => {
                        event.stopPropagation()
                        setCollapsed(collapsed => !collapsed)
                    }}
                    aria-label={collapsed ? 'Show filter description' : 'Hide filter description'}
                >
                    <small className="text-monospace">i</small>
                    <CollapseIcon className="icon-inline" />
                </button>
            </span>
            <Collapse isOpen={!collapsed}>
                <div className={styles.description}>
                    {searchReference.description && (
                        <Markdown dangerousInnerHTML={renderMarkdown(searchReference.description)} />
                    )}
                    {searchReference.alias && (
                        <p>
                            Alias:{' '}
                            <span className="text-code search-filter-keyword">
                                {searchReference.alias}
                                {isFilterInfo(searchReference) ? ':' : ''}
                            </span>
                        </p>
                    )}
                    {isFilterInfo(searchReference) && isNegatableFilter(searchReference.field) && (
                        <p>
                            Negation: <span className="test-code search-filter-keyword">-{searchReference.field}:</span>
                            {searchReference.alias && (
                                <>
                                    {' '}
                                    | <span className="test-code search-filter-keyword">-{searchReference.alias}:</span>
                                </>
                            )}
                            <br />
                            <span className={styles.placeholder}>(opt + click filter in reference list)</span>
                        </p>
                    )}
                    {searchReference.examples && (
                        <>
                            <div className="font-weight-medium">Examples</div>
                            <div className={classNames('text-code', styles.examples)}>
                                {searchReference.examples.map(example => (
                                    <p key={example}>
                                        <SearchReferenceExample example={example} onClick={onExampleClick} />
                                    </p>
                                ))}
                            </div>
                        </>
                    )}
                </div>
            </Collapse>
        </li>
    )
}

interface FilterInfoListProps<T extends SearchReferenceInfo> {
    filters: T[]
    onClick: (info: T, negate: boolean) => void
    onExampleClick: (example: string) => void
}

const FilterInfoList = ({ filters, onClick, onExampleClick }: FilterInfoListProps<FilterInfo>): ReactElement => (
    <ul className={styles.list}>
        {filters.map(filterInfo => (
            <SearchReferenceEntry
                key={filterInfo.field + filterInfo.value}
                searchReference={filterInfo}
                onClick={onClick}
                onExampleClick={onExampleClick}
            />
        ))}
    </ul>
)

export interface SearchReferenceProps
    extends Omit<PatternTypeProps, 'setPatternType'>,
        Omit<CaseSensitivityProps, 'setCaseSensitivity'>,
        VersionContextProps,
        TelemetryProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    query: string
    filter: string
    navbarSearchQueryState: QueryState
    onNavbarQueryChange: (queryState: QueryState) => void
    isSourcegraphDotCom: boolean
}

const SearchReference = (props: SearchReferenceProps): ReactElement => {
    const [selectedTab, setSelectedTab] = useLocalStorage(SEARCH_REFERENCE_TAB_KEY, 0)

    const { onNavbarQueryChange, navbarSearchQueryState, telemetryService } = props
    const filter = props.filter.trim()
    const hasFilter = filter.length > 0

    const selectedFilters = useMemo(() => {
        if (!hasFilter) {
            return filterInfos
        }
        const searchTerms = parseSearchInput(filter)
        return filterInfos.filter(info => matches(searchTerms, info))
    }, [filter, hasFilter])

    const updateQuery = useCallback(
        (searchReference: FilterInfo, negate: boolean) => {
            const updatedQuery = updateQueryWithFilterAndExample(
                navbarSearchQueryState.query,
                searchReference.field,
                searchReference,
                {
                    singular: Boolean(FILTERS[searchReference.field].singular),
                    negate: negate && isNegatableFilter(searchReference.field),
                    emptyValue: shouldShowSuggestions(searchReference),
                }
            )
            onNavbarQueryChange({
                changeSource: QueryChangeSource.searchReference,
                query: updatedQuery.query,
                selectionRange: updatedQuery.placeholderRange,
                revealRange: updatedQuery.filterRange,
                showSuggestions: shouldShowSuggestions(searchReference),
            })
        },
        [onNavbarQueryChange, navbarSearchQueryState]
    )
    const updateQueryWithOperator = useCallback(
        (info: OperatorInfo) => {
            onNavbarQueryChange({
                query: navbarSearchQueryState.query + ` ${info.operator} `,
            })
        },
        [onNavbarQueryChange, navbarSearchQueryState]
    )
    const updateQueryWithExample = useCallback(
        (example: string) => {
            telemetryService.log(hasFilter ? 'SearchReferenceSearchedAndClicked' : 'SearchReferenceFilterClicked')
            onNavbarQueryChange({ query: navbarSearchQueryState.query.trimEnd() + ' ' + example })
        },
        [onNavbarQueryChange, navbarSearchQueryState, hasFilter, telemetryService]
    )

    const filterList = (
        <FilterInfoList filters={selectedFilters} onClick={updateQuery} onExampleClick={updateQueryWithExample} />
    )

    return (
        <div>
            {hasFilter ? (
                filterList
            ) : (
                <Tabs index={selectedTab} onChange={setSelectedTab}>
                    <TabList className={styles.tablist}>
                        <Tab>Common</Tab>
                        <Tab>All filters</Tab>
                        <Tab>Operators</Tab>
                    </TabList>
                    <TabPanels>
                        <TabPanel>
                            <FilterInfoList
                                filters={commonFilters}
                                onClick={updateQuery}
                                onExampleClick={updateQueryWithExample}
                            />
                        </TabPanel>
                        <TabPanel>{filterList}</TabPanel>
                        <TabPanel>
                            <ul className={styles.list}>
                                {operatorInfo.map(operatorInfo => (
                                    <SearchReferenceEntry
                                        searchReference={operatorInfo}
                                        key={operatorInfo.operator + operatorInfo.value}
                                        onClick={updateQueryWithOperator}
                                        onExampleClick={updateQueryWithExample}
                                    />
                                ))}
                            </ul>
                        </TabPanel>
                    </TabPanels>
                </Tabs>
            )}
            <p className={styles.footer}>
                <small>
                    <Link target="blank" to="https://docs.sourcegraph.com/code_search/reference/queries">
                        Search syntax <ExternalLinkIcon className="icon-inline" />
                    </Link>
                </small>
            </p>
        </div>
    )
}

export function getSearchReferenceFactory(
    props: Omit<SearchReferenceProps, 'filter'>
): (filter: string) => ReactElement {
    return (filter: string) => <SearchReference {...props} filter={filter} />
}
