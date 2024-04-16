import { Observable, concatMap, from, map } from 'rxjs'

import { fetchBlameHunksMemoized, type BlameHunkData } from '@sourcegraph/web/src/repo/blame/shared'

import { getGraphQLClient, mapOrThrow } from '$lib/graphql'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { BlobDiffQuery, BlobPageQuery, BlobSyntaxHighlightQuery } from './page.gql'

export const load: PageLoad = ({ parent, params, url }) => {
    const revisionToCompare = url.searchParams.get('rev')
    const client = getGraphQLClient()
    const { repoName, revision = '' } = parseRepoRevision(params.repo)
    const resolvedRevision = resolveRevision(parent, revision)
    const isBlame = url.searchParams.get('view') === 'blame'

    var blameData: Observable<BlameHunkData> | undefined = undefined
    if (isBlame) {
        const blameHunks = from(resolvedRevision).pipe(
            concatMap(resolvedRevision =>
                fetchBlameHunksMemoized({ repoName, revision: resolvedRevision, filePath: params.path })
            )
        )

        blameData = from(parent()).pipe(
            concatMap(({ resolvedRevision }) =>
                blameHunks.pipe(
                    map(blameHunks => ({
                        externalURLs: resolvedRevision.repo.externalURLs,
                        current: blameHunks,
                    }))
                )
            )
        )
    }

    return {
        graphQLClient: client,
        filePath: params.path,
        blob: resolvedRevision
            .then(resolvedRevision =>
                client.query(BlobPageQuery, {
                    repoName,
                    revspec: resolvedRevision,
                    path: params.path,
                })
            )
            .then(mapOrThrow(result => result.data?.repository?.commit?.blob ?? null)),
        highlights: resolvedRevision
            .then(resolvedRevision =>
                client.query(BlobSyntaxHighlightQuery, {
                    repoName,
                    revspec: resolvedRevision,
                    path: params.path,
                    disableTimeout: false,
                })
            )
            .then(mapOrThrow(result => result.data?.repository?.commit?.blob?.highlight.lsif ?? '')),
        compare: revisionToCompare
            ? {
                  revisionToCompare,
                  diff: client
                      .query(BlobDiffQuery, {
                          repoName,
                          revspec: revisionToCompare,
                          paths: [params.path],
                      })
                      .then(mapOrThrow(result => result.data?.repository?.commit?.diff.fileDiffs.nodes[0] ?? null)),
              }
            : null,
        blameData,
    }
}
