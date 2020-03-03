import { Filter } from './parser'
import { SearchSuggestion } from '../../graphql/schema'
import {
    FilterType,
    isNegatedFilter,
    resolveNegatedFilter,
    NegatableFilter,
    isNegatableFilter,
    isFilterType,
} from '../interactive/util'
import { Omit } from 'utility-types'

interface BaseFilterDefinition {
    alias?: string
    description: string
    discreteValues?: string[]
    suggestions?: SearchSuggestion['__typename'] | string[]
    default?: string
    /** Whether the filter may only be used 0 or 1 times in a query. */
    singular?: boolean
}

interface NegatableFilterDefinition extends Omit<BaseFilterDefinition, 'description'> {
    negatable: true
    description: (negated: boolean) => string
}

export type FilterDefinition = BaseFilterDefinition | NegatableFilterDefinition

const LANGUAGES: string[] = [
    'c',
    'cpp',
    'csharp',
    'css',
    'go',
    'graphql',
    'haskell',
    'html',
    'java',
    'javascript',
    'json',
    'lua',
    'markdown',
    'php',
    'powershell',
    'python',
    'r',
    'ruby',
    'sass',
    'swift',
    'typescript',
]

export const FILTERS: Record<NegatableFilter, NegatableFilterDefinition> &
    Record<Exclude<FilterType, NegatableFilter>, BaseFilterDefinition> = {
    [FilterType.after]: {
        description: 'Commits made after a certain date',
    },
    [FilterType.archived]: {
        description: 'Include results from archived repositories.',
        singular: true,
    },
    [FilterType.author]: {
        description: 'The author of a commit',
    },
    [FilterType.before]: {
        description: 'Commits made before a certain date',
    },
    [FilterType.case]: {
        description: 'Treat the search pattern as case-sensitive.',
        discreteValues: ['yes', 'no'],
        default: 'no',
        singular: true,
    },
    [FilterType.content]: {
        description:
            'Explicitly overrides the search pattern. Used for explicitly delineating the search pattern to search for in case of clashes.',
        singular: true,
    },
    [FilterType.count]: {
        description: 'Number of results to fetch (integer)',
        singular: true,
    },
    [FilterType.file]: {
        alias: 'f',
        negatable: true,
        description: negated =>
            `${negated ? 'Exclude' : 'Include only'} results from files matching the given regex pattern.`,
        suggestions: 'File',
    },
    [FilterType.fork]: {
        discreteValues: ['yes', 'no', 'only'],
        description: 'Include results from forked repositories.',
        singular: true,
    },
    [FilterType.lang]: {
        negatable: true,
        description: negated => `${negated ? 'Exclude' : 'Include only'} results from the given language`,
        suggestions: LANGUAGES,
    },
    [FilterType.message]: {
        description: 'Commits with messages matching a certain string',
    },
    [FilterType.patterntype]: {
        discreteValues: ['regexp', 'literal', 'structural'],
        description: 'The pattern type (regexp, literal, structural) in use',
        singular: true,
    },
    [FilterType.repo]: {
        alias: 'r',
        negatable: true,
        description: negated =>
            `${negated ? 'Exclude' : 'Include only'} results from repositories matching the given regex pattern.`,
        suggestions: 'Repository',
    },
    [FilterType.repogroup]: {
        description: 'group-name (include results from the named group)',
        singular: true,
    },
    [FilterType.repohascommitafter]: {
        description: '"string specifying time frame" (filter out stale repositories without recent commits)',
        singular: true,
    },
    [FilterType.repohasfile]: {
        negatable: true,
        description: negated =>
            `${negated ? 'Exclude' : 'Include only'} results from repos that contain a matching file`,
    },
    [FilterType.timeout]: {
        description: 'Duration before timeout',
        singular: true,
    },
    [FilterType.type]: {
        description: 'Limit results to the specified type.',
        discreteValues: ['diff', 'commit', 'symbol', 'repo', 'path'],
    },
}

/**
 * Returns the {@link FilterDefinition} for the given filterType if it exists, or `undefined` otherwise.
 */
export const resolveFilter = (
    filterType: string
):
    | { type: NegatableFilter; negated: boolean; definition: NegatableFilterDefinition }
    | { type: Exclude<FilterType, NegatableFilter>; definition: BaseFilterDefinition }
    | undefined => {
    filterType = filterType.toLowerCase()
    if (isNegatedFilter(filterType)) {
        const type = resolveNegatedFilter(filterType)
        return {
            type,
            definition: FILTERS[type],
            negated: true,
        }
    }
    if (isFilterType(filterType)) {
        if (isNegatableFilter(filterType)) {
            return {
                type: filterType,
                definition: FILTERS[filterType],
                negated: false,
            }
        }
        if (FILTERS[filterType]) {
            return { type: filterType, definition: FILTERS[filterType] }
        }
    }
    for (const [type, definition] of Object.entries(FILTERS as Record<FilterType, FilterDefinition>)) {
        if (definition.alias && filterType === definition.alias) {
            return {
                type: type as Exclude<FilterType, NegatableFilter>,
                definition: definition as BaseFilterDefinition,
            }
        }
    }
    return undefined
}

/**
 * Validates a filter given its type and value.
 */
export const validateFilter = (
    filterType: string,
    filterValue: Filter['filterValue']
): { valid: true } | { valid: false; reason: string } => {
    const typeAndDefinition = resolveFilter(filterType)
    if (!typeAndDefinition) {
        return { valid: false, reason: 'Invalid filter type.' }
    }
    const { definition } = typeAndDefinition
    if (
        definition.discreteValues &&
        (!filterValue ||
            filterValue.token.type !== 'literal' ||
            !definition.discreteValues.includes(filterValue.token.value))
    ) {
        return {
            valid: false,
            reason: `Invalid filter value, expected one of: ${definition.discreteValues.join(', ')}.`,
        }
    }
    return { valid: true }
}

/** Whether a given filter type may only be used 0 or 1 times in a query. */
export const isSingularFilter = (filter: string): boolean =>
    Object.keys(FILTERS)
        .filter(key => FILTERS[key as FilterType].singular)
        .includes(filter)
