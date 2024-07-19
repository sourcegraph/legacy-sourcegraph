import { error } from '@sveltejs/kit'

import type { ResolvedRevision } from '$lib/repo/utils'
import { RevisionNotFoundError } from '$lib/shared'

import type { LayoutLoad } from './$types'

export const load: LayoutLoad = async ({ parent }) => {
    // By validating the resolved revision here we can guarantee to
    // subpages that if they load the requested revision exists. This
    // relieves subpages from testing whether the revision is valid.
    const { revision, defaultBranch, resolvedRepository } = await parent()

    const commit = resolvedRepository.commit || resolvedRepository.changelist?.commit

    if (!commit) {
        error(404, new RevisionNotFoundError(revision))
    }

    return {
        resolvedRevision: {
            repo: resolvedRepository,
            commitID: commit.oid,
            defaultBranch,
        } satisfies ResolvedRevision,
    }
}
