import { Base64 } from 'js-base64'

import * as sourcegraph from './api'

export const linkURL = 'https://docs.sourcegraph.com/code_intelligence/explanations/precise_code_intelligence'

// Aggregable badges
//
// These indicators are the new hotness, and are tagged on each result that we send
// back. These will be displayed aggregated at the file level to determine result
// precision.

export const preciseBadge: sourcegraph.AggregableBadge = {
    text: 'precise',
    linkURL,
    hoverMessage: "This data comes from a pre-computed code intelligence index of this project's source code.",
}

export const searchBasedBadge: sourcegraph.AggregableBadge = {
    text: 'search-based',
    linkURL,
    hoverMessage: 'This data is generated by a heuristic text-based search.',
}

export const partialHoverNoDefinitionBadge: sourcegraph.AggregableBadge = {
    text: 'partially precise',
    linkURL,
    hoverMessage:
        'It looks like this symbol is defined in another repository that does not have a pre-computed code intelligence index. Go to definition may be imprecise.',
}

export const partialDefinitionNoHoverBadge: sourcegraph.AggregableBadge = {
    text: 'partially precise',
    linkURL,
    hoverMessage:
        'It looks like this symbol is defined in another repository that does not have a pre-computed code intelligence index. This hover text may be imprecise.',
}

// Hover Alerts
//
// These alerts are displayed within the hover overlay and are emitted only for the
// first result (to avoid duplication). These alerts are generally dismissible and
// show only redundant information that is also present in the aggregable badges.
// We send more data than necessary here to support an old format that only affects
// instances older than 3.26.
//
// Once we no longer support this old format, we can remove the extra field and
// collapse some identically-rendered alerts (tooltip text and links are ignored).

const makeAlert = ({
    message,
    hoverMessage,
    type,
}: {
    message: string
    hoverMessage?: string
    type?: string
}): sourcegraph.Badged<sourcegraph.HoverAlert> => {
    const legacyFields = {
        badge: { kind: 'info' as const, linkURL, hoverMessage },
    }

    return {
        type,
        iconKind: 'info',
        summary: {
            kind: sourcegraph.MarkupKind.Markdown,
            value: `${message}<br /> <a href="${linkURL}" target="_blank">Learn more about precise code intelligence</a>`,
        },
        ...legacyFields,
    }
}

export const lsif = makeAlert({
    type: 'LSIFAvailableNoCaveat',
    message: 'Precise result.',
    hoverMessage:
        "This data comes from a pre-computed code intelligence index of this project's source code. Click to learn how to add this capability to all of your projects!",
})

export const lsifPartialHoverOnly = makeAlert({
    type: 'LSIFAvailableNoCaveat',
    message: 'Partially precise result.',
    hoverMessage:
        'It looks like this symbol is defined in another repository that does not have a pre-computed code intelligence index. Click to learn how to make these results precise by enabling precise indexing for that project.',
})

export const lsifPartialDefinitionOnly = makeAlert({
    type: 'LSIFAvailableNoCaveat',
    message: 'Partially precise result.',
    hoverMessage:
        'It looks like this symbol is defined in another repository that does not have a pre-computed code intelligence index. Click to learn how to make these results precise by enabling precise indexing for that project.',
})

export const searchLSIFSupportRobust = makeAlert({
    type: 'SearchResultLSIFSupportRobust',
    message: 'Search-based result.',
    hoverMessage:
        'This data is generated by a heuristic text-based search. Click to learn how to make these results precise by enabling precise indexing for this project.',
})

export const searchLSIFSupportExperimental = makeAlert({
    type: 'SearchResultExperimentalLSIFSupport',
    message: 'Search-based result.',
    hoverMessage:
        "This data is generated by a heuristic text-based search. Existing LSIF indexers for this language aren't totally robust yet, but you can click here to learn how to give them a try.",
})

export const searchLSIFSupportNone = makeAlert({
    type: 'SearchResultNoLSIFSupport',
    message: 'Search-based result.',
})

// Badge indicators
//
// These indicators were deprecated in 3.26, but we still need to send them back
// from the extensions as we don't know how old the instance we're interfacing
// with is and they might not have the code to display the new indicators.

const rawInfoIcon = (color: string): string => `
    <svg xmlns='http://www.w3.org/2000/svg' style="width:24px;height:24px" viewBox="0 0 24 24" fill="${color}">
        <path d="
            M11,
            9H13V7H11M12,
            20C7.59,
            20 4,
            16.41 4,
            12C4,
            7.59 7.59,
            4 12,
            4C16.41,
            4 20,
            7.59 20,
            12C20,
            16.41 16.41,
            20 12,
            20M12,
            2A10,
            10 0 0,
            0 2,
            12A10,
            10 0 0,
            0 12,
            22A10,
            10 0 0,
            0 22,
            12A10,
            10 0 0,
            0 12,
            2M11,
            17H13V11H11V17Z"
        />
    </svg>
`

const infoIcon = (color: string): string =>
    `data:image/svg+xml;base64,${Base64.encode(
        rawInfoIcon(color)
            .split('\n')
            .map(line => line.trimStart())
            .join(' ')
    )}`

/**
 * This badge is placed on all results that come from search-based providers.
 */
export const impreciseBadge = {
    kind: 'info',
    icon: infoIcon('#ffffff'),
    light: { icon: infoIcon('#000000') },
    hoverMessage:
        'Search-based results - click to see how these results are calculated and how to get precise intelligence with LSIF.',
    linkURL: 'https://docs.sourcegraph.com/code_intelligence/explanations/basic_code_intelligence',
}
