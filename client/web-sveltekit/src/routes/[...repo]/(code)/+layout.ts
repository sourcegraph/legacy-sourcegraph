import { dirname } from 'path'

import { browser } from '$app/environment'
import { isErrorLike } from '$lib/common'
import { fetchRepoCommits } from '$lib/repo/api/commits'
import { fetchSidebarFileTree } from '$lib/repo/api/tree'

import type { LayoutLoad } from './$types'

// Signifies the path of the repository root
const REPO_ROOT = '.'

let getRootPath = (_repo: string, path: string) => path

// We keep state in the browser to load the tree entries of the "highest" directory that was visited.
if (browser) {
    let topTreePath: Record<string, string> = {}

    getRootPath = (repo: string, path: string) => {
        const treePath = topTreePath[repo]
        if (treePath && (treePath === REPO_ROOT || path.startsWith(treePath))) {
            return topTreePath[repo]
        }
        return (topTreePath[repo] = path)
    }
}

export const load: LayoutLoad = async ({ parent, params }) => {
    const { resolvedRevision: resolvedRevisionOrError, repoName, revision } = await parent()
    const parentPath = getRootPath(repoName, params.path ? dirname(params.path) : REPO_ROOT)
    const resolvedRevision = isErrorLike(resolvedRevisionOrError) ? null : resolvedRevisionOrError

    return {
        parentPath,
        deferred: {
            // Fetches the most recent commits for current blob, tree or repo root
            codeCommits: resolvedRevision
                ? fetchRepoCommits({
                      repoID: resolvedRevision.repo.id,
                      revision: resolvedRevision.commitID,
                      filePath: params.path ?? null,
                  })
                : null,
            fileTree: resolvedRevision
                ? fetchSidebarFileTree({
                      repoName,
                      commitID: resolvedRevision.commitID,
                      revision: revision ?? '',
                      filePath: parentPath,
                  })
                : null,
        },
    }
}
