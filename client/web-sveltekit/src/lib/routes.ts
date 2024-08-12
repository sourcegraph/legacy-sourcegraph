// This file is automatically generated. Do not edit it directly.
// To regenerate this file, run 'bazel run //client/web-sveltekit:write_generated'.
export interface SvelteKitRoute {
    // The SvelteKit route ID
    id: string
    // The regular expression pattern that matches the corresponding path
    pattern: RegExp
    // Whether the route is the repository root
    isRepoRoot: boolean
}

// prettier-ignore
export const svelteKitRoutes: SvelteKitRoute[] = [
    {
        id: '/[...repo=reporev]/(validrev)/(code)',
        pattern: new RegExp('^/(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,})))(?:@(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,}))))?/?$'),
        isRepoRoot: true,
    },
    {
        id: '/[...repo=reporev]/(validrev)/(code)/-/blob/[...path]',
        pattern: new RegExp('^/(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,})))(?:@(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,}))))?/-/blob(?:/.*)?/?$'),
        isRepoRoot: false,
    },
    {
        id: '/[...repo=reporev]/(validrev)/(code)/-/tree/[...path]',
        pattern: new RegExp('^/(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,})))(?:@(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,}))))?/-/tree(?:/.*)?/?$'),
        isRepoRoot: false,
    },
    {
        id: '/[...repo=reporev]/(validrev)/-/commits/[...path]',
        pattern: new RegExp('^/(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,})))(?:@(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,}))))?/-/commits(?:/.*)?/?$'),
        isRepoRoot: false,
    },
    {
        id: '/[...repo=reporev]/-/branches',
        pattern: new RegExp('^/(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,})))(?:@(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,}))))?/-/branches/?$'),
        isRepoRoot: false,
    },
    {
        id: '/[...repo=reporev]/-/branches/all',
        pattern: new RegExp('^/(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,})))(?:@(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,}))))?/-/branches/all/?$'),
        isRepoRoot: false,
    },
    {
        id: '/[...repo=reporev]/-/commit/[...revspec]',
        pattern: new RegExp('^/(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,})))(?:@(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,}))))?/-/commit(?:/.*)?/?$'),
        isRepoRoot: false,
    },
    {
        id: '/[...repo=reporev]/-/compare/[...rangeSpec]',
        pattern: new RegExp('^/(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,})))(?:@(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,}))))?/-/compare(?:/.*)?/?$'),
        isRepoRoot: false,
    },
    {
        id: '/[...repo=reporev]/-/stats/contributors',
        pattern: new RegExp('^/(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,})))(?:@(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,}))))?/-/stats/contributors/?$'),
        isRepoRoot: false,
    },
    {
        id: '/[...repo=reporev]/-/tags',
        pattern: new RegExp('^/(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,})))(?:@(?:(?:(?:[^@/-]|(?:[^/@]{2,}))/)*(?:[^@/-]|(?:[^/@]{2,}))))?/-/tags/?$'),
        isRepoRoot: false,
    },
    {
        id: '/[community=communitySearchContext]',
        pattern: new RegExp('^/(backstage|chakraui|cncf|julia|kubernetes|o3de|stackstorm|stanford|temporal)/?$'),
        isRepoRoot: false,
    },
    {
        id: '/search',
        pattern: new RegExp('^/search/?$'),
        isRepoRoot: false,
    },
    
]
