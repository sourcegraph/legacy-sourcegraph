import { ProxyMarked, proxyMarker } from '@sourcegraph/comlink'
import * as clientType from '@sourcegraph/extension-api-types'
import { Subject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'

/** @internal */
export interface ExtRootsAPI extends ProxyMarked {
    $acceptRoots(roots: readonly clientType.WorkspaceRoot[]): void
}

/** @internal */
export class ExtRoots implements ExtRootsAPI, ProxyMarked {
    public readonly [proxyMarker] = true

    private roots: readonly sourcegraph.WorkspaceRoot[] = []

    /**
     * Returns a single workspace root by URI.
     *
     * @internal
     */
    public getByUri(uri: string): sourcegraph.WorkspaceRoot {
        // Not efficient for many workspaces, but currently we only support one workspace.
        const root = this.roots.find(root => root.uri.href === uri)
        if (!root) {
            throw new Error(`Workspace root with URI ${uri} does not exist`)
        }
        return root
    }

    /**
     * Returns all workspace roots.
     *
     * @internal
     */
    public getAll(): readonly sourcegraph.WorkspaceRoot[] {
        return this.roots
    }

    public readonly changes = new Subject<void>()

    public $acceptRoots(roots: clientType.WorkspaceRoot[]): void {
        this.roots = Object.freeze(roots.map(plain => ({ ...plain, uri: new URL(plain.uri) })))
        this.changes.next()
    }
}
