package lsifstore

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// ErrNotFound occurs when data does not exist for a requested bundle.
var ErrNotFound = errors.New("data does not exist")

// Exists determines if the path exists in the database.
func (s *Store) Exists(ctx context.Context, bundleID int, path string) (_ bool, err error) {
	ctx, endObservation := s.operations.exists.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
	}})
	defer endObservation(1, observation.Args{})

	_, exists, err := basestore.ScanFirstString(s.Store.Query(ctx, sqlf.Sprintf(existsQuery, bundleID, path)))
	return exists, err
}

const existsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:Exists
SELECT path FROM lsif_data_documents WHERE dump_id = %s AND path = %s LIMIT 1
`

// Ranges returns definition, reference, and hover data for each range within the given span of lines.
func (s *Store) Ranges(ctx context.Context, bundleID int, path string, startLine, endLine int) (_ []CodeIntelligenceRange, err error) {
	ctx, traceLog, endObservation := s.operations.ranges.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("startLine", startLine),
		log.Int("endLine", endLine),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(documentQuery, bundleID, path)))
	if err != nil || !exists {
		return nil, err
	}
	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))

	ranges := map[ID]RangeData{}
	for id, r := range documentData.Document.Ranges {
		if RangeIntersectsSpan(r, startLine, endLine) {
			ranges[id] = r
		}
	}
	traceLog(log.Int("numIntersectingRanges", len(ranges)))

	resultIDMap := make(map[ID]struct{}, 2*len(ranges))
	for _, r := range ranges {
		if r.DefinitionResultID != "" {
			resultIDMap[r.DefinitionResultID] = struct{}{}
		}
		if r.ReferenceResultID != "" {
			resultIDMap[r.ReferenceResultID] = struct{}{}
		}
	}

	resultIDs := make([]ID, 0, len(resultIDMap))
	for id := range resultIDMap {
		resultIDs = append(resultIDs, id)
	}

	locations, err := s.locations(ctx, bundleID, resultIDs)
	if err != nil {
		return nil, err
	}

	codeintelRanges := make([]CodeIntelligenceRange, 0, len(ranges))
	for _, r := range ranges {
		var hoverText string
		if r.HoverResultID != "" {
			if text, exists := documentData.Document.HoverResults[r.HoverResultID]; exists {
				hoverText = text
			}
		}

		// Return only references that are in the same file. Otherwise this set
		// gets very big and such results are of limited use to consumers such as
		// the code intel extensions, which only use references for highlighting
		// uses of an identifier within the same file.
		fileLocalReferences := make([]Location, 0, len(locations[r.ReferenceResultID]))
		for _, r := range locations[r.ReferenceResultID] {
			if r.Path == path {
				fileLocalReferences = append(fileLocalReferences, r)
			}
		}

		codeintelRanges = append(codeintelRanges, CodeIntelligenceRange{
			Range:       newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
			Definitions: locations[r.DefinitionResultID],
			References:  fileLocalReferences,
			HoverText:   hoverText,
		})
	}

	sort.Slice(codeintelRanges, func(i, j int) bool {
		return compareBundleRanges(codeintelRanges[i].Range, codeintelRanges[j].Range)
	})

	return codeintelRanges, nil
}

const documentQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:{Ranges,Definitions,References,Hover,MonikersByPosition}
SELECT dump_id, path, data FROM lsif_data_documents WHERE dump_id = %s AND path = %s LIMIT 1
`

// Definitions returns the set of locations defining the symbol at the given position.
func (s *Store) Definitions(ctx context.Context, bundleID int, path string, line, character int) (_ []Location, err error) {
	ctx, traceLog, endObservation := s.operations.definitions.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(documentQuery, bundleID, path)))
	if err != nil || !exists {
		return nil, err
	}

	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))
	ranges := FindRanges(documentData.Document.Ranges, line, character)
	traceLog(log.Int("numIntersectingRanges", len(ranges)))

	orderedResultIDs := extractResultIDs(ranges, func(r RangeData) ID { return r.DefinitionResultID })
	locationsMap, err := s.locations(ctx, bundleID, orderedResultIDs)
	if err != nil {
		return nil, err
	}

	totalCount := 0
	for _, locations := range locationsMap {
		totalCount += len(locations)
	}
	traceLog(log.Int("totalCount", totalCount))

	for _, resultID := range orderedResultIDs {
		if locations := locationsMap[resultID]; len(locations) > 0 {
			return locations, nil
		}
	}

	return nil, nil
}

// PagedReferences returns the set of locations referencing the symbol at the given position.
func (s *Store) PagedReferences(ctx context.Context, bundleID int, path string, line, character, limit, offset int) (_ []Location, _ int, err error) {
	ctx, traceLog, endObservation := s.operations.pagedReferences.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(documentQuery, bundleID, path)))
	if err != nil || !exists {
		return nil, 0, err
	}

	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))
	ranges := FindRanges(documentData.Document.Ranges, line, character)
	traceLog(log.Int("numIntersectingRanges", len(ranges)))

	orderedResultIDs := extractResultIDs(ranges, func(r RangeData) ID { return r.ReferenceResultID })
	locationsMap, err := s.locations(ctx, bundleID, orderedResultIDs)
	if err != nil {
		return nil, 0, err
	}

	totalCount := 0
	for _, locations := range locationsMap {
		totalCount += len(locations)
	}
	traceLog(log.Int("totalCount", totalCount))

	max := totalCount
	if totalCount > limit {
		max = limit
	}

	locations := make([]Location, 0, max)
outer:
	for _, resultID := range orderedResultIDs {
		for _, location := range locationsMap[resultID] {
			offset--
			if offset >= 0 {
				continue
			}

			locations = append(locations, location)
			if len(locations) >= limit {
				break outer
			}
		}
	}

	return locations, totalCount, nil
}

// Hover returns the hover text of the symbol at the given position.
func (s *Store) Hover(ctx context.Context, bundleID int, path string, line, character int) (_ string, _ Range, _ bool, err error) {
	ctx, traceLog, endObservation := s.operations.hover.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(documentQuery, bundleID, path)))
	if err != nil || !exists {
		return "", Range{}, false, err
	}

	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))
	ranges := FindRanges(documentData.Document.Ranges, line, character)
	traceLog(log.Int("numIntersectingRanges", len(ranges)))

	for _, r := range ranges {
		if text, ok := documentData.Document.HoverResults[r.HoverResultID]; ok {
			return text, newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter), true, nil
		}
	}

	return "", Range{}, false, nil
}

// Diagnostics returns the diagnostics for the documents that have the given path prefix. This method
// also returns the size of the complete result set to aid in pagination.
func (s *Store) Diagnostics(ctx context.Context, bundleID int, prefix string, limit, offset int) (_ []Diagnostic, _ int, err error) {
	ctx, traceLog, endObservation := s.operations.diagnostics.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("prefix", prefix),
		log.Int("limit", limit),
		log.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	documentData, err := s.scanDocumentData(s.Store.Query(ctx, sqlf.Sprintf(diagnosticsQuery, bundleID, prefix+"%")))
	if err != nil {
		return nil, 0, err
	}
	traceLog(log.Int("numDocuments", len(documentData)))

	totalCount := 0
	for _, documentData := range documentData {
		totalCount += len(documentData.Document.Diagnostics)
	}
	traceLog(log.Int("totalCount", totalCount))

	diagnostics := make([]Diagnostic, 0, limit)
	for _, documentData := range documentData {
		for _, diagnostic := range documentData.Document.Diagnostics {
			offset--

			if offset < 0 && len(diagnostics) < limit {
				diagnostics = append(diagnostics, Diagnostic{
					DumpID:         bundleID,
					Path:           documentData.Path,
					DiagnosticData: diagnostic,
				})
			}
		}
	}

	return diagnostics, totalCount, nil
}

const diagnosticsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:Diagnostics
SELECT dump_id, path, data FROM lsif_data_documents WHERE dump_id = %s AND path LIKE %s ORDER BY path
`

// MonikersByPosition returns all monikers attached ranges containing the given position. If multiple
// ranges contain the position, then this method will return multiple sets of monikers. Each slice
// of monikers are attached to a single range. The order of the output slice is "outside-in", so that
// the range attached to earlier monikers enclose the range attached to later monikers.
func (s *Store) MonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) (_ [][]MonikerData, err error) {
	ctx, traceLog, endObservation := s.operations.monikersByPosition.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(documentQuery, bundleID, path)))
	if err != nil || !exists {
		return nil, err
	}

	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))
	ranges := FindRanges(documentData.Document.Ranges, line, character)
	traceLog(log.Int("numIntersectingRanges", len(ranges)))

	var monikerData [][]MonikerData
	for _, r := range ranges {
		var batch []MonikerData
		for _, monikerID := range r.MonikerIDs {
			if moniker, exists := documentData.Document.Monikers[monikerID]; exists {
				batch = append(batch, moniker)
			}
		}
		traceLog(log.Int("numMonikersForRange", len(batch)))

		monikerData = append(monikerData, batch)
	}
	traceLog(log.Int("numMonikers", len(monikerData)))

	return monikerData, nil
}

// BulkMonikerResults returns the locations within one of the given bundles that define or reference
// one of the given monikers. This method also returns the size of the complete result set to aid in
// pagination.
func (s *Store) BulkMonikerResults(ctx context.Context, tableName string, uploadIDs []int, monikers []MonikerData, limit, offset int) (_ []Location, _ int, err error) {
	ctx, traceLog, endObservation := s.operations.bulkMonikerResults.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("tableName", tableName),
		log.Int("numUploadIDs", len(uploadIDs)),
		log.String("uploadIDs", intsToString(uploadIDs)),
		log.Int("numMonikers", len(monikers)),
		log.String("monikers", monikersToString(monikers)),
		log.Int("limit", limit),
		log.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	if len(uploadIDs) == 0 || len(monikers) == 0 {
		return nil, 0, nil
	}

	idQueries := make([]*sqlf.Query, 0, len(uploadIDs))
	for _, id := range uploadIDs {
		idQueries = append(idQueries, sqlf.Sprintf("%s", id))
	}

	monikerQueries := make([]*sqlf.Query, 0, len(monikers))
	for _, arg := range monikers {
		monikerQueries = append(monikerQueries, sqlf.Sprintf("(%s, %s)", arg.Scheme, arg.Identifier))
	}

	locationData, err := s.scanQualifiedMonikerLocations(s.Store.Query(ctx, sqlf.Sprintf(
		bulkMonikerResultsQuery,
		sqlf.Sprintf(fmt.Sprintf("lsif_data_%s", tableName)),
		sqlf.Join(idQueries, ", "),
		sqlf.Join(monikerQueries, ", "),
	)))
	if err != nil {
		return nil, 0, err
	}

	totalCount := 0
	for _, monikerLocations := range locationData {
		totalCount += len(monikerLocations.Locations)
	}
	traceLog(
		log.Int("numDumps", len(locationData)),
		log.Int("totalCount", totalCount),
	)

	max := totalCount
	if totalCount > limit {
		max = limit
	}

	locations := make([]Location, 0, max)
outer:
	for _, monikerLocations := range locationData {
		for _, row := range monikerLocations.Locations {
			offset--
			if offset >= 0 {
				continue
			}

			locations = append(locations, Location{
				DumpID: monikerLocations.DumpID,
				Path:   row.URI,
				Range:  newRange(row.StartLine, row.StartCharacter, row.EndLine, row.EndCharacter),
			})

			if len(locations) >= limit {
				break outer
			}
		}
	}
	traceLog(log.Int("numLocations", len(locations)))

	return locations, totalCount, nil
}

const bulkMonikerResultsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:BulkMonikerResults
SELECT dump_id, scheme, identifier, data FROM %s WHERE dump_id IN (%s) AND (scheme, identifier) IN (%s) ORDER BY (scheme, identifier, dump_id)
`

// PackageInformation looks up package information data by identifier.
func (s *Store) PackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ PackageInformationData, _ bool, err error) {
	ctx, endObservation := s.operations.packageInformation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.String("packageInformationID", packageInformationID),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(packageInformationQuery, bundleID, path)))
	if err != nil || !exists {
		return PackageInformationData{}, false, err
	}

	packageInformationData, exists := documentData.Document.PackageInformation[ID(packageInformationID)]
	return packageInformationData, exists, nil
}

const packageInformationQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:PackageInformation
SELECT dump_id, path, data FROM lsif_data_documents WHERE dump_id = %s AND path = %s LIMIT 1
`

var ErrNoMetadata = errors.New("no rows in meta table")

// locations returns the locations for the given definition or reference identifiers.
func (s *Store) locations(ctx context.Context, bundleID int, ids []ID) (_ map[ID][]Location, err error) {
	ctx, traceLog, endObservation := s.operations.locations.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.Int("numIDs", len(ids)),
		log.String("ids", idsToString(ids)),
	}})
	defer endObservation(1, observation.Args{})

	if len(ids) == 0 {
		return nil, nil
	}

	numResultChunks, exists, err := basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(locationsMetaQuery, bundleID)))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNoMetadata
	}

	resultChunkIndexMap := map[int]struct{}{}
	for _, id := range ids {
		resultChunkIndexMap[HashKey(id, numResultChunks)] = struct{}{}
	}
	indexes := make([]int, 0, len(resultChunkIndexMap))
	for index := range resultChunkIndexMap {
		indexes = append(indexes, index)
	}
	traceLog(
		log.Int("numIndexes", len(indexes)),
		log.String("indexes", intsToString(indexes)),
	)

	indexQueries := make([]*sqlf.Query, 0, len(indexes))
	for _, index := range indexes {
		indexQueries = append(indexQueries, sqlf.Sprintf("%s", index))
	}

	resultChunkData, err := s.scanQualifiedResultChunkData(s.Store.Query(ctx, sqlf.Sprintf(locationsResultChunkQuery, bundleID, sqlf.Join(indexQueries, ","))))
	if err != nil {
		return nil, err
	}

	resultChunksByIndex := map[int]ResultChunkData{}
	for _, resultChunkData := range resultChunkData {
		resultChunksByIndex[resultChunkData.Index] = resultChunkData.ResultChunk
	}

	locationsByResultID := map[ID]map[string][]ID{}
	for _, id := range ids {
		resultChunkData := resultChunksByIndex[HashKey(id, numResultChunks)]

		documentIDRangeIDs, exists := resultChunkData.DocumentIDRangeIDs[id]
		if !exists {
			continue
		}

		resultData := map[string][]ID{}
		for _, documentIDRangeID := range documentIDRangeIDs {
			path, ok := resultChunkData.DocumentPaths[documentIDRangeID.DocumentID]
			if !ok {
				continue
			}

			resultData[path] = append(resultData[path], documentIDRangeID.RangeID)
		}

		locationsByResultID[id] = resultData
	}

	var paths []string
	for _, locations := range locationsByResultID {
		for path := range locations {
			paths = append(paths, path)
		}
	}
	traceLog(
		log.Int("numPaths", len(paths)),
		log.String("paths", strings.Join(paths, ", ")),
	)

	pathQueries := make([]*sqlf.Query, 0, len(paths))
	for _, path := range paths {
		pathQueries = append(pathQueries, sqlf.Sprintf("%s", path))
	}
	if len(pathQueries) == 0 {
		return nil, nil
	}

	documentData, err := s.scanDocumentData(s.Store.Query(ctx, sqlf.Sprintf(locationsDocumentQuery, bundleID, sqlf.Join(pathQueries, ","))))
	if err != nil {
		return nil, err
	}

	documentsByPath := make(map[string]DocumentData, len(documentData))
	for _, documentData := range documentData {
		documentsByPath[documentData.Path] = documentData.Document
	}

	totalCount := 0
	locationsByID := map[ID][]Location{}
	for _, id := range ids {
		var locations []Location
		for path, rangeIDs := range locationsByResultID[id] {
			for _, rangeID := range rangeIDs {
				r, exists := documentsByPath[path].Ranges[rangeID]
				if !exists {
					continue
				}

				locations = append(locations, Location{
					DumpID: bundleID,
					Path:   path,
					Range:  newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter),
				})
			}
		}

		sort.Slice(locations, func(i, j int) bool {
			if locations[i].Path == locations[j].Path {
				return compareBundleRanges(locations[i].Range, locations[j].Range)
			}

			return strings.Compare(locations[i].Path, locations[j].Path) < 0
		})

		locationsByID[id] = locations
		totalCount += len(locations)
		traceLog(log.Int("numLocationsForID", len(locations)))
	}
	traceLog(log.Int("numLocations", totalCount))

	return locationsByID, nil
}

const locationsMetaQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:locations
SELECT num_result_chunks FROM lsif_data_metadata WHERE dump_id = %s
`

const locationsResultChunkQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:locations
SELECT dump_id, idx, data FROM lsif_data_result_chunks WHERE dump_id = %s AND idx IN (%s)
`

const locationsDocumentQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/bundle.go:locations
SELECT dump_id, path, data FROM lsif_data_documents WHERE dump_id = %s AND path IN (%s)
`

// compareBundleRanges returns true if r1's start position occurs before r2's start position.
func compareBundleRanges(r1, r2 Range) bool {
	cmp := r1.Start.Line - r2.Start.Line
	if cmp == 0 {
		cmp = r1.Start.Character - r2.Start.Character
	}

	return cmp < 0
}

func newRange(startLine, startCharacter, endLine, endCharacter int) Range {
	return Range{
		Start: Position{
			Line:      startLine,
			Character: startCharacter,
		},
		End: Position{
			Line:      endLine,
			Character: endCharacter,
		},
	}
}

// extractResultIDs extracts result identifiers from each range in the given list.
// The output list is relative to the input range list, but with duplicates removed.
func extractResultIDs(ranges []RangeData, fn func(r RangeData) ID) []ID {
	resultIDs := make([]ID, 0, len(ranges))
	resultIDMap := make(map[ID]struct{}, len(ranges))

	for _, r := range ranges {
		resultID := fn(r)

		if _, ok := resultIDMap[resultID]; !ok && resultID != "" {
			resultIDs = append(resultIDs, resultID)
			resultIDMap[resultID] = struct{}{}
		}
	}

	return resultIDs
}

func monikersToString(vs []MonikerData) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, fmt.Sprintf("%s:%s", v.Scheme, v.Identifier))
	}

	return strings.Join(strs, ", ")
}

func idsToString(vs []ID) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, string(v))
	}

	return strings.Join(strs, ", ")
}
