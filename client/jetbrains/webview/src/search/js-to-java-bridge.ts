import { encode } from 'js-base64'

import { splitPath } from '@sourcegraph/shared/src/components/RepoLink'
import {
    ContentMatch,
    getRepoMatchUrl,
    PathMatch,
    SearchMatch,
    SearchType,
    SymbolMatch,
} from '@sourcegraph/shared/src/search/stream'

import { loadContent } from './lib/blob'
import { PluginConfig, Search, Theme } from './types'

export interface PreviewContent {
    timeAsISOString: string
    resultType: SearchType
    fileName?: string
    repoUrl: string
    path?: string
    content: string | null
    symbolName?: string
    symbolContainerName?: string
    commitMessagePreview?: string
    lineNumber?: number
    absoluteOffsetAndLengths?: number[][]
    relativeUrl?: string
}

export interface PreviewLoadingRequest {
    action: 'previewLoading'
    arguments: { timeAsISOString: string }
}

export interface PreviewRequest {
    action: 'preview'
    arguments: PreviewContent
}

interface ClearPreviewRequest {
    action: 'clearPreview'
    arguments: { timeAsISOString: string }
}

interface OpenRequest {
    action: 'open'
    arguments: PreviewContent
}

interface GetConfigRequest {
    action: 'getConfig'
}

interface GetThemeRequest {
    action: 'getTheme'
}

export interface SaveLastSearchRequest {
    action: 'saveLastSearch'
    arguments: Search
}

interface LoadLastSearchRequest {
    action: 'loadLastSearch'
}

interface IndicateFinishedLoadingRequest {
    action: 'indicateFinishedLoading'
}

export type Request =
    | PreviewLoadingRequest
    | PreviewRequest
    | OpenRequest
    | GetConfigRequest
    | GetThemeRequest
    | SaveLastSearchRequest
    | LoadLastSearchRequest
    | ClearPreviewRequest
    | IndicateFinishedLoadingRequest

let lastPreviewUpdateCallSendDateTime = new Date()

export async function getConfigAlwaysFulfill(): Promise<PluginConfig> {
    try {
        return (await callJava({ action: 'getConfig' })) as PluginConfig
    } catch (error) {
        console.error(`Failed to get config: ${(error as Error).message}`)
        return {
            instanceURL: 'https://sourcegraph.com',
            isGlobbingEnabled: false,
            accessToken: null,
        }
    }
}

export async function getThemeAlwaysFulfill(): Promise<Theme> {
    try {
        return (await callJava({ action: 'getTheme' })) as Theme
    } catch (error) {
        console.error(`Failed to get theme: ${(error as Error).message}`)
        return {
            isDarkTheme: false,
            intelliJTheme: {},
            syntaxTheme: {},
        }
    }
}

export async function indicateFinishedLoading(): Promise<void> {
    try {
        await callJava({ action: 'indicateFinishedLoading' })
    } catch (error) {
        console.error(`Failed to indicate “finished loading”: ${(error as Error).message}`)
    }
}

export async function onPreviewChange(match: SearchMatch, lineOrSymbolMatchIndex?: number): Promise<void> {
    try {
        const initiationDateTime = new Date()
        if (match.type === 'content' || match.type === 'path' || match.type === 'symbol') {
            lastPreviewUpdateCallSendDateTime = initiationDateTime
            await callJava({
                action: 'previewLoading',
                arguments: { timeAsISOString: lastPreviewUpdateCallSendDateTime.toISOString() },
            })
        }
        const previewContent = await createPreviewContent(match, lineOrSymbolMatchIndex)
        if (initiationDateTime < lastPreviewUpdateCallSendDateTime) {
            // Apparently, the content was slow to load, and we already sent a newer request in the meantime.
            // The best we can do is to ignore this change to prevent overwriting the newer content.
            return
        }
        await callJava({ action: 'preview', arguments: previewContent })
    } catch (error) {
        console.error(`Failed to preview match: ${(error as Error).message}`)
    }
}

export async function onPreviewClear(): Promise<void> {
    try {
        lastPreviewUpdateCallSendDateTime = new Date()
        await callJava({
            action: 'clearPreview',
            arguments: { timeAsISOString: lastPreviewUpdateCallSendDateTime.toISOString() },
        })
    } catch (error) {
        console.error(`Failed to clear preview: ${(error as Error).message}`)
    }
}

export async function onOpen(match: SearchMatch, lineOrSymbolMatchIndex?: number): Promise<void> {
    try {
        await callJava({ action: 'open', arguments: await createPreviewContent(match, lineOrSymbolMatchIndex) })
    } catch (error) {
        console.error(`Failed to open match: ${(error as Error).message}`)
    }
}

export async function loadLastSearchAlwaysFulfill(): Promise<Search | null> {
    try {
        return (await callJava({ action: 'loadLastSearch' })) as Search
    } catch (error) {
        console.error(`Failed to get last search: ${(error as Error).message}`)
        return null
    }
}

export function saveLastSearch(lastSearch: Search): void {
    callJava({ action: 'saveLastSearch', arguments: lastSearch })
        .then(() => {
            console.log(`Saved last search: ${JSON.stringify(lastSearch)}`)
        })
        .catch((error: Error) => {
            console.error(`Failed to save last search: ${error.message}`)
        })
}

async function callJava(request: Request): Promise<object> {
    return window.callJava(request)
}

export async function createPreviewContent(
    match: SearchMatch,
    lineOrSymbolMatchIndex: number | undefined
): Promise<PreviewContent> {
    if (match.type === 'commit') {
        const isCommitResult = match.content.startsWith('```COMMIT_EDITMSG')
        const content = prepareContent(
            isCommitResult
                ? match.content.replace(/^```COMMIT_EDITMSG\n([\S\s]*)\n```$/, '$1')
                : match.content.replace(/^```diff\n([\S\s]*)\n```$/, '$1')
        )
        return {
            timeAsISOString: new Date().toISOString(),
            resultType: isCommitResult ? 'commit' : 'diff',
            repoUrl: match.repository,
            content,
            commitMessagePreview: match.message.split('\n', 1)[0],
            relativeUrl: match.url,
        }
    }

    if (match.type === 'content') {
        return createPreviewContentForContentMatch(match, lineOrSymbolMatchIndex as number)
    }

    if (match.type === 'path') {
        return createPreviewContentForPathMatch(match)
    }

    if (match.type === 'repo') {
        return {
            timeAsISOString: new Date().toISOString(),
            resultType: match.type,
            repoUrl: getRepoMatchUrl(match).slice(1),
            content: null,
            relativeUrl: getRepoMatchUrl(match),
        }
    }

    if (match.type === 'symbol') {
        return createPreviewContentForSymbolMatch(match, lineOrSymbolMatchIndex as number)
    }

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore This is here in preparation for future match types
    console.log(`Unknown match type: “${match.type}”`)

    return {
        timeAsISOString: new Date().toISOString(),
        resultType: null,
        repoUrl: '',
        content: null,
    }
}

async function createPreviewContentForContentMatch(
    match: ContentMatch,
    lineMatchIndex: number
): Promise<PreviewContent> {
    const fileName = splitPath(match.path)[1]
    const content = await loadContent(match)
    const characterCountUntilLine = getCharacterCountUntilLine(content, match.lineMatches[lineMatchIndex].lineNumber)
    const absoluteOffsetAndLengths = getAbsoluteOffsetAndLengths(
        match.lineMatches[lineMatchIndex].offsetAndLengths,
        characterCountUntilLine
    )

    return {
        timeAsISOString: new Date().toISOString(),
        resultType: 'file',
        fileName,
        repoUrl: match.repository,
        path: match.path,
        content: prepareContent(content),
        lineNumber: match.lineMatches[lineMatchIndex].lineNumber,
        absoluteOffsetAndLengths,
    }
}

async function createPreviewContentForPathMatch(match: PathMatch): Promise<PreviewContent> {
    const fileName = splitPath(match.path)[1]
    const content = await loadContent(match)

    return {
        timeAsISOString: new Date().toISOString(),
        resultType: match.type,
        fileName,
        repoUrl: match.repository,
        path: match.path,
        content: prepareContent(content),
    }
}

async function createPreviewContentForSymbolMatch(
    match: SymbolMatch,
    symbolMatchIndex: number
): Promise<PreviewContent> {
    const fileName = splitPath(match.path)[1]
    const content = await loadContent(match)
    const symbolMatch = match.symbols[symbolMatchIndex]

    console.log(symbolMatch)

    return {
        timeAsISOString: new Date().toISOString(),
        resultType: match.type,
        fileName,
        repoUrl: match.repository,
        path: match.path,
        content: prepareContent(content),
        symbolName: symbolMatch.name,
        symbolContainerName: symbolMatch.containerName,
        lineNumber: getLineFromSourcegraphUrl(symbolMatch.url),
        absoluteOffsetAndLengths: getAbsoluteOffsetAndLengthsFromSourcegraphUrl(symbolMatch.url, content),
        relativeUrl: '',
    }
}

// We encode the content as base64-encoded string to avoid encoding errors in the Java JSON parser.
// The Java side also does not expect `\r\n` line endings, so we replace them with `\n`.
//
// We can not use the native btoa() function because it does not support all Unicode characters.
function prepareContent(content: string | null): string | null {
    if (content === null) {
        return null
    }
    return encode(content.replaceAll('\r\n', '\n'))
}

// NOTE: This might be slow when the content is a really large file and the match is in the
// beginning of the file because we convert all rows to an array first.
//
// If we ever run into issues with large files, this is a place to get some wins.
function getCharacterCountUntilLine(content: string | null, lineNumber: number): number {
    if (content === null) {
        return 0
    }

    let count = 0
    const lines = content.replaceAll('\r\n', '\n').split('\n')
    for (let index = 0; index < lineNumber; index++) {
        count += lines[index].length + 1
    }
    return count
}

function getAbsoluteOffsetAndLengths(offsetAndLengths: number[][], characterCountUntilLine: number): number[][] {
    return offsetAndLengths.map(offsetAndLength => [offsetAndLength[0] + characterCountUntilLine, offsetAndLength[1]])
}

function getLineFromSourcegraphUrl(url: string): number {
    const offsets = extractStartAndEndOffsetsFromSourcegraphUrl(url)
    if (offsets === null) {
        return -1
    }
    return offsets.start.line
}

function getAbsoluteOffsetAndLengthsFromSourcegraphUrl(url: string, content: string | null): number[][] {
    const offsets = extractStartAndEndOffsetsFromSourcegraphUrl(url)
    if (offsets === null) {
        return []
    }
    const absoluteStart = getCharacterCountUntilLine(content, offsets.start.line) + offsets.start.col
    const absoluteEnd = getCharacterCountUntilLine(content, offsets.end.line) + offsets.end.col
    return [[absoluteStart, absoluteEnd - absoluteStart]]
}

// Parses a Sourcegraph URL and extracts the offsets from it. E.g.:
//
//     /github.com/apache/kafka/-/blob/streams/src/main/j…ls/graph/SourceGraphNode.java?L28:23-28:38
//
// Will be parsed into:
//
//    {
//      start: {
//        line: 28,
//        column: 23
//      },
//      end: {
//         line: 28,
//         column: 38
//       }
//    }
function extractStartAndEndOffsetsFromSourcegraphUrl(
    url: string
): null | { start: { line: number; col: number }; end: { line: number; col: number } } {
    const match = url.match(/L(\d+):(\d+)-(\d+):(\d+)$/)
    if (match === null) {
        return null
    }
    return {
        start: { line: parseInt(match[1], 10) - 1, col: parseInt(match[2], 10) - 1 },
        end: { line: parseInt(match[3], 10) - 1, col: parseInt(match[4], 10) - 1 },
    }
}
