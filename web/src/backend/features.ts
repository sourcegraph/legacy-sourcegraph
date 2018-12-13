import { HoverMerged } from '@sourcegraph/codeintellify/lib/types'
import { Location } from '@sourcegraph/extension-api-types'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import {
    FileSpec,
    parseRepoURI,
    PositionSpec,
    RepoSpec,
    ResolvedRevSpec,
    RevSpec,
    toPrettyBlobURL,
} from '../../../shared/src/util/url'
import { toAbsoluteBlobURL } from '../util/url'

/**
 * Specifies an LSP mode.
 */
export interface ModeSpec {
    /** The LSP mode, which identifies the language server to use. */
    mode: string
}

export { HoverMerged } // reexport to avoid needing to change all import sites - TODO(sqs): actually go change all them

/**
 * Fetches hover information for the given location.
 *
 * @param ctx the location
 * @return hover for the location
 */
export function getHover(
    ctx: RepoSpec & ResolvedRevSpec & FileSpec & PositionSpec,
    { extensionsController }: ExtensionsControllerProps
): Observable<HoverMerged | null> {
    return extensionsController.services.textDocumentHover
        .getHover({
            textDocument: { uri: `git://${ctx.repoName}?${ctx.commitID}#${ctx.filePath}` },
            position: {
                character: ctx.position.character - 1,
                line: ctx.position.line - 1,
            },
        })
        .pipe(map(hover => hover as HoverMerged | null))
}

/**
 * Fetches definitions (in the same repository) for the given location.
 *
 * @param ctx the location
 * @return definitions of the symbol at the location
 */
function getDefinition(
    ctx: RepoSpec & ResolvedRevSpec & FileSpec & PositionSpec,
    { extensionsController }: ExtensionsControllerProps
): Observable<Location[] | null> {
    return extensionsController.services.textDocumentDefinition.getLocations({
        textDocument: { uri: `git://${ctx.repoName}?${ctx.commitID}#${ctx.filePath}` },
        position: {
            character: ctx.position.character - 1,
            line: ctx.position.line - 1,
        },
    })
}

/**
 * Fetches the destination URL for the "Go to definition" action in the hover.
 *
 * Only the first URL is returned, even if there are results from multiple providers or a provider returns
 * multiple results.
 *
 * @param ctx the location containing the token whose definition to jump to
 * @return destination URL
 */
export function getJumpURL(
    ctx: RepoSpec & RevSpec & ResolvedRevSpec & FileSpec & PositionSpec,
    extensions: ExtensionsControllerProps
): Observable<string | null> {
    return getDefinition(ctx, extensions).pipe(
        map(defs => {
            if (!defs || defs.length === 0) {
                return null
            }
            const def = defs[0]

            const uri = parseRepoURI(def.uri) as RepoSpec & RevSpec & ResolvedRevSpec & FileSpec & PositionSpec
            if (def.range) {
                uri.position = { line: def.range.start.line + 1, character: def.range.start.character + 1 }
            }
            if (uri.repoName === ctx.repoName && uri.commitID === ctx.commitID) {
                // Use pretty rev from the current context for same-repo J2D.
                uri.rev = ctx.rev
                return toPrettyBlobURL(uri)
            }
            return toAbsoluteBlobURL(uri)
        })
    )
}
