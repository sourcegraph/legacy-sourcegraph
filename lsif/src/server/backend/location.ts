import * as lsp from 'vscode-languageserver-protocol'
import * as pgModels from '../../shared/models/pg'
import { OrderedSet } from '../../shared/datastructures/orderedset'

/** A location with the identifier of the dump that contains it. */
export interface InternalLocation {
    dumpId: pgModels.DumpId
    path: string
    range: lsp.Range
}

/** A location with the dump that contains it. */
export interface ResolvedInternalLocation {
    dump: pgModels.LsifDump
    path: string
    range: lsp.Range
}

/** A duplicate-free list of locations ordered by time of insertion. */
export class OrderedLocationSet extends OrderedSet<InternalLocation> {
    /**
     * Create a new ordered locations set.
     *
     * @param values A set of values used to seed the set.
     * @param trusted Whether the given values are already deduplicated.
     */
    constructor(locations?: InternalLocation[], trusted = false) {
        super(
            (location: InternalLocation): string =>
                [
                    location.dumpId,
                    location.path,
                    location.range.start.line,
                    location.range.start.character,
                    location.range.end.line,
                    location.range.end.character,
                ].join(':'),
            locations,
            trusted
        )
    }
}
