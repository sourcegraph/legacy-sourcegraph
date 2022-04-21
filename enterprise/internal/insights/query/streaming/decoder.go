package streaming

import (
	"fmt"

	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

// TabulationDecoder will tabulate the result counts per repository.
func TabulationDecoder() (streamhttp.FrontendStreamDecoder, *int, map[string]*SearchMatch, []string, []string) {
	var totalCount int
	repoCounts := make(map[string]*SearchMatch)
	addCount := func(repo string, repoId int32, count int) {
		if forRepo, ok := repoCounts[repo]; !ok {
			repoCounts[repo] = &SearchMatch{
				RepositoryID:   repoId,
				RepositoryName: repo,
				MatchCount:     count,
			}
			return
		} else {
			forRepo.MatchCount += count
		}
	}
	var errors []string
	var skippedReasons []string

	return streamhttp.FrontendStreamDecoder{
		OnProgress: func(progress *streamapi.Progress) {
			if !progress.Done {
				return
			}
			// Skipped elements are built progressively for a Progress update until it is Done, so
			// we want to register its contents only once it is done.
			for _, skipped := range progress.Skipped {
				skippedReasons = append(skippedReasons, fmt.Sprintf("%s: %s", skipped.Reason, skipped.Message))
			}
		},
		OnMatches: func(matches []streamhttp.EventMatch) {
			for _, match := range matches {
				switch match := match.(type) {
				case *streamhttp.EventContentMatch:
					count := 0
					for _, lineMatch := range match.LineMatches {
						count += len(lineMatch.OffsetAndLengths)
					}
					totalCount += count
					addCount(match.Repository, match.RepositoryID, count)
				case *streamhttp.EventPathMatch:
					totalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventRepoMatch:
					totalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventCommitMatch:
					totalCount += 1
					addCount(match.Repository, match.RepositoryID, 1)
				case *streamhttp.EventSymbolMatch:
					count := len(match.Symbols)
					totalCount += count
					addCount(match.Repository, match.RepositoryID, count)
				}
			}
		},
		OnError: func(eventError *streamhttp.EventError) {
			errors = append(errors, eventError.Message)
		},
	}, &totalCount, repoCounts, skippedReasons, errors
}

type SearchMatch struct {
	RepositoryID   int32
	RepositoryName string
	MatchCount     int
}
