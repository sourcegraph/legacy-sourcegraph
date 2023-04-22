package embeddings

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type EmbeddingIndex struct {
	Embeddings      []int8
	ColumnDimension int
	RowMetadata     []RepoEmbeddingRowMetadata
	Ranks           []float32
}

// Row returns the embeddings for the nth row in the index
func (index *EmbeddingIndex) Row(n int) []int8 {
	return index.Embeddings[n*index.ColumnDimension : (n+1)*index.ColumnDimension]
}

type RepoEmbeddingRowMetadata struct {
	FileName  string `json:"fileName"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
}

type RepoEmbeddingIndex struct {
	RepoName  api.RepoName
	Revision  api.CommitID
	CodeIndex EmbeddingIndex
	TextIndex EmbeddingIndex
}

type ContextDetectionEmbeddingIndex struct {
	MessagesWithAdditionalContextMeanEmbedding    []float32
	MessagesWithoutAdditionalContextMeanEmbedding []float32
}

type EmbeddingSearchResults struct {
	CodeResults []EmbeddingSearchResult `json:"codeResults"`
	TextResults []EmbeddingSearchResult `json:"textResults"`
}

type EmbeddingSearchResult struct {
	RepoEmbeddingRowMetadata
	// The row number in the index to correlate this result back with its source.
	RowNum  int    `json:"rowNum"`
	Content string `json:"content"`
	// Experimental: Clients should not rely on any particular format of debug
	Debug string `json:"debug,omitempty"`
}

// DEPRECATED: to support decoding old indexes, we need a struct
// we can decode into directly. This struct is the same shape
// as the old indexes and should not be changed without migrating
// all existing indexes to the new format.
type OldRepoEmbeddingIndex struct {
	RepoName  api.RepoName
	Revision  api.CommitID
	CodeIndex OldEmbeddingIndex
	TextIndex OldEmbeddingIndex
}

func (o *OldRepoEmbeddingIndex) ToNewIndex() *RepoEmbeddingIndex {
	return &RepoEmbeddingIndex{
		RepoName:  o.RepoName,
		Revision:  o.Revision,
		CodeIndex: o.CodeIndex.ToNewIndex(),
		TextIndex: o.TextIndex.ToNewIndex(),
	}
}

type OldEmbeddingIndex struct {
	Embeddings      []float32
	ColumnDimension int
	RowMetadata     []RepoEmbeddingRowMetadata
	Ranks           []float32
}

func (o *OldEmbeddingIndex) ToNewIndex() EmbeddingIndex {
	return EmbeddingIndex{
		Embeddings:      Quantize(o.Embeddings),
		ColumnDimension: o.ColumnDimension,
		RowMetadata:     o.RowMetadata,
		Ranks:           o.Ranks,
	}
}

type EmbedRepoStats struct {
	// Repo name
	RepoName api.RepoName
	Revision api.CommitID

	HasRanks       bool
	InputFileCount int

	CodeIndexStats EmbedFilesStats
	TextIndexStats EmbedFilesStats
}

type EmbedFilesStats struct {
	InputFileCount int
	Dimensions     int

	// Options
	NoSplitTokensThreshold         int
	ChunkTokensThreshold           int
	ChunkEarlySplitTokensThreshold int
	MaxEmbeddingVectors            int

	Duration               time.Duration
	HitMaxEmbeddingVectors bool
	SkippedReasons         map[string]int
}
