// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.29.1
// 	protoc        (unknown)
// source: searcher.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// SearchRequest is set of parameters for a search.
type SearchRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// repo is the name of the repo to search (e.g. "github.com/gorilla/mux")
	Repo string `protobuf:"bytes,1,opt,name=repo,proto3" json:"repo,omitempty"`
	// repo_id is the Sourcegraph repository ID of the repo to search
	RepoId uint32 `protobuf:"varint,2,opt,name=repo_id,json=repoId,proto3" json:"repo_id,omitempty"`
	// commit_oid is the 40-character commit hash for the commit to be searched.
	// It is required to be resolved, not a ref like HEAD or master.
	CommitOid string `protobuf:"bytes,3,opt,name=commit_oid,json=commitOid,proto3" json:"commit_oid,omitempty"`
	// indexed is whether the revision to be searched is indexed or
	// unindexed. This matters for structural search because it will query
	// Zoekt for indexed structural search.
	Indexed     bool         `protobuf:"varint,4,opt,name=indexed,proto3" json:"indexed,omitempty"`
	PatternInfo *PatternInfo `protobuf:"bytes,5,opt,name=pattern_info,json=patternInfo,proto3" json:"pattern_info,omitempty"`
	// URL specifies the repository's Git remote URL (for gitserver). It is
	// optional. See (gitserver.ExecRequest).URL for documentation on what it is
	// used for.
	Url string `protobuf:"bytes,6,opt,name=url,proto3" json:"url,omitempty"`
	// branch is used for structural search as an alternative to Commit
	// because Zoekt only takes branch names
	Branch string `protobuf:"bytes,7,opt,name=branch,proto3" json:"branch,omitempty"`
	// fetch_timeout is the amount of time to wait for a repo archive to
	// fetch.
	//
	// This timeout should be low when searching across many repos so that
	// unfetched repos don't delay the search, and because we are likely
	// to get results from the repos that have already been fetched.
	//
	// This timeout should be high when searching across a single repo
	// because returning results slowly is better than returning no
	// results at all.
	//
	// This only times out how long we wait for the fetch request; the
	// fetch will still happen in the background so future requests don't
	// have to wait.
	FetchTimeout *durationpb.Duration `protobuf:"bytes,8,opt,name=fetch_timeout,json=fetchTimeout,proto3" json:"fetch_timeout,omitempty"`
	// num_context_lines is the number of additional lines of context
	// (before and after the matched lines) to return with the match.
	NumContextLines int32 `protobuf:"varint,10,opt,name=num_context_lines,json=numContextLines,proto3" json:"num_context_lines,omitempty"`
}

func (x *SearchRequest) Reset() {
	*x = SearchRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_searcher_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchRequest) ProtoMessage() {}

func (x *SearchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_searcher_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchRequest.ProtoReflect.Descriptor instead.
func (*SearchRequest) Descriptor() ([]byte, []int) {
	return file_searcher_proto_rawDescGZIP(), []int{0}
}

func (x *SearchRequest) GetRepo() string {
	if x != nil {
		return x.Repo
	}
	return ""
}

func (x *SearchRequest) GetRepoId() uint32 {
	if x != nil {
		return x.RepoId
	}
	return 0
}

func (x *SearchRequest) GetCommitOid() string {
	if x != nil {
		return x.CommitOid
	}
	return ""
}

func (x *SearchRequest) GetIndexed() bool {
	if x != nil {
		return x.Indexed
	}
	return false
}

func (x *SearchRequest) GetPatternInfo() *PatternInfo {
	if x != nil {
		return x.PatternInfo
	}
	return nil
}

func (x *SearchRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *SearchRequest) GetBranch() string {
	if x != nil {
		return x.Branch
	}
	return ""
}

func (x *SearchRequest) GetFetchTimeout() *durationpb.Duration {
	if x != nil {
		return x.FetchTimeout
	}
	return nil
}

func (x *SearchRequest) GetNumContextLines() int32 {
	if x != nil {
		return x.NumContextLines
	}
	return 0
}

// SearchResponse is a message in the response stream for Search
type SearchResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Message:
	//
	//	*SearchResponse_FileMatch
	//	*SearchResponse_DoneMessage
	Message isSearchResponse_Message `protobuf_oneof:"message"`
}

func (x *SearchResponse) Reset() {
	*x = SearchResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_searcher_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchResponse) ProtoMessage() {}

func (x *SearchResponse) ProtoReflect() protoreflect.Message {
	mi := &file_searcher_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchResponse.ProtoReflect.Descriptor instead.
func (*SearchResponse) Descriptor() ([]byte, []int) {
	return file_searcher_proto_rawDescGZIP(), []int{1}
}

func (m *SearchResponse) GetMessage() isSearchResponse_Message {
	if m != nil {
		return m.Message
	}
	return nil
}

func (x *SearchResponse) GetFileMatch() *FileMatch {
	if x, ok := x.GetMessage().(*SearchResponse_FileMatch); ok {
		return x.FileMatch
	}
	return nil
}

func (x *SearchResponse) GetDoneMessage() *SearchResponse_Done {
	if x, ok := x.GetMessage().(*SearchResponse_DoneMessage); ok {
		return x.DoneMessage
	}
	return nil
}

type isSearchResponse_Message interface {
	isSearchResponse_Message()
}

type SearchResponse_FileMatch struct {
	FileMatch *FileMatch `protobuf:"bytes,1,opt,name=file_match,json=fileMatch,proto3,oneof"`
}

type SearchResponse_DoneMessage struct {
	DoneMessage *SearchResponse_Done `protobuf:"bytes,2,opt,name=done_message,json=doneMessage,proto3,oneof"`
}

func (*SearchResponse_FileMatch) isSearchResponse_Message() {}

func (*SearchResponse_DoneMessage) isSearchResponse_Message() {}

// FileMatch is a file that matched the search query along
// with the parts of the file that matched.
type FileMatch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The file's path
	Path []byte `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// A list of matched chunks
	ChunkMatches []*ChunkMatch `protobuf:"bytes,2,rep,name=chunk_matches,json=chunkMatches,proto3" json:"chunk_matches,omitempty"`
	// Whether the limit was hit while searching this
	// file. Indicates that the results for this file
	// may not be complete.
	LimitHit bool `protobuf:"varint,3,opt,name=limit_hit,json=limitHit,proto3" json:"limit_hit,omitempty"`
}

func (x *FileMatch) Reset() {
	*x = FileMatch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_searcher_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileMatch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileMatch) ProtoMessage() {}

func (x *FileMatch) ProtoReflect() protoreflect.Message {
	mi := &file_searcher_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileMatch.ProtoReflect.Descriptor instead.
func (*FileMatch) Descriptor() ([]byte, []int) {
	return file_searcher_proto_rawDescGZIP(), []int{2}
}

func (x *FileMatch) GetPath() []byte {
	if x != nil {
		return x.Path
	}
	return nil
}

func (x *FileMatch) GetChunkMatches() []*ChunkMatch {
	if x != nil {
		return x.ChunkMatches
	}
	return nil
}

func (x *FileMatch) GetLimitHit() bool {
	if x != nil {
		return x.LimitHit
	}
	return false
}

// ChunkMatch is a matched chunk of a file.
type ChunkMatch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The raw content that contains the match. Will always
	// contain complete lines.
	Content []byte `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
	// The location relative to the start of the file
	// where the chunk content starts.
	ContentStart *Location `protobuf:"bytes,2,opt,name=content_start,json=contentStart,proto3" json:"content_start,omitempty"`
	// A list of ranges within the chunk content that match
	// the search query.
	Ranges []*Range `protobuf:"bytes,3,rep,name=ranges,proto3" json:"ranges,omitempty"`
}

func (x *ChunkMatch) Reset() {
	*x = ChunkMatch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_searcher_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChunkMatch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChunkMatch) ProtoMessage() {}

func (x *ChunkMatch) ProtoReflect() protoreflect.Message {
	mi := &file_searcher_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChunkMatch.ProtoReflect.Descriptor instead.
func (*ChunkMatch) Descriptor() ([]byte, []int) {
	return file_searcher_proto_rawDescGZIP(), []int{3}
}

func (x *ChunkMatch) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

func (x *ChunkMatch) GetContentStart() *Location {
	if x != nil {
		return x.ContentStart
	}
	return nil
}

func (x *ChunkMatch) GetRanges() []*Range {
	if x != nil {
		return x.Ranges
	}
	return nil
}

type Range struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start *Location `protobuf:"bytes,1,opt,name=start,proto3" json:"start,omitempty"`
	End   *Location `protobuf:"bytes,2,opt,name=end,proto3" json:"end,omitempty"`
}

func (x *Range) Reset() {
	*x = Range{}
	if protoimpl.UnsafeEnabled {
		mi := &file_searcher_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Range) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Range) ProtoMessage() {}

func (x *Range) ProtoReflect() protoreflect.Message {
	mi := &file_searcher_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Range.ProtoReflect.Descriptor instead.
func (*Range) Descriptor() ([]byte, []int) {
	return file_searcher_proto_rawDescGZIP(), []int{4}
}

func (x *Range) GetStart() *Location {
	if x != nil {
		return x.Start
	}
	return nil
}

func (x *Range) GetEnd() *Location {
	if x != nil {
		return x.End
	}
	return nil
}

// A location represents an offset within a file.
type Location struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The byte offset from the beginning of the byte slice.
	Offset int32 `protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	// The number of newlines in the file before the offset.
	Line int32 `protobuf:"varint,2,opt,name=line,proto3" json:"line,omitempty"`
	// The rune offset from the beginning of the last line.
	Column int32 `protobuf:"varint,3,opt,name=column,proto3" json:"column,omitempty"`
}

func (x *Location) Reset() {
	*x = Location{}
	if protoimpl.UnsafeEnabled {
		mi := &file_searcher_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Location) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Location) ProtoMessage() {}

func (x *Location) ProtoReflect() protoreflect.Message {
	mi := &file_searcher_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Location.ProtoReflect.Descriptor instead.
func (*Location) Descriptor() ([]byte, []int) {
	return file_searcher_proto_rawDescGZIP(), []int{5}
}

func (x *Location) GetOffset() int32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *Location) GetLine() int32 {
	if x != nil {
		return x.Line
	}
	return 0
}

func (x *Location) GetColumn() int32 {
	if x != nil {
		return x.Column
	}
	return 0
}

type PatternInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// pattern is the search query. It is a regular expression if IsRegExp
	// is true, otherwise a fixed string. eg "route variable"
	Pattern string `protobuf:"bytes,1,opt,name=pattern,proto3" json:"pattern,omitempty"`
	// is_negated if true will invert the matching logic for regexp searches.
	// IsNegated=true is not supported for structural searches.
	IsNegated bool `protobuf:"varint,2,opt,name=is_negated,json=isNegated,proto3" json:"is_negated,omitempty"`
	// is_regexp if true will treat the pattern as a regular expression.
	IsRegexp bool `protobuf:"varint,3,opt,name=is_regexp,json=isRegexp,proto3" json:"is_regexp,omitempty"`
	// is_structural if true will treat the pattern as a Comby structural search
	// pattern.
	IsStructural bool `protobuf:"varint,4,opt,name=is_structural,json=isStructural,proto3" json:"is_structural,omitempty"`
	// is_case_sensitive if false will ignore the case of text and pattern
	// when finding matches.
	IsCaseSensitive bool `protobuf:"varint,6,opt,name=is_case_sensitive,json=isCaseSensitive,proto3" json:"is_case_sensitive,omitempty"`
	// exclude_pattern is a pattern that may not match the returned files' paths.
	// eg '**/node_modules'
	ExcludePattern string `protobuf:"bytes,7,opt,name=exclude_pattern,json=excludePattern,proto3" json:"exclude_pattern,omitempty"`
	// include_patterns is a list of patterns that must *all* match the returned
	// files' paths.
	// eg '**/node_modules'
	//
	// The patterns are ANDed together; a file's path must match all patterns
	// for it to be kept. That is also why it is a list (unlike the singular
	// ExcludePattern); it is not possible in general to construct a single
	// glob or Go regexp that represents multiple such patterns ANDed together.
	IncludePatterns []string `protobuf:"bytes,8,rep,name=include_patterns,json=includePatterns,proto3" json:"include_patterns,omitempty"`
	// path_patterns_are_case_sensitive indicates that exclude_pattern and
	// include_patterns are case sensitive.
	PathPatternsAreCaseSensitive bool `protobuf:"varint,9,opt,name=path_patterns_are_case_sensitive,json=pathPatternsAreCaseSensitive,proto3" json:"path_patterns_are_case_sensitive,omitempty"`
	// limit is the cap on the total number of matches returned.
	// A match is either a path match, or a fragment of a line matched by the
	// query.
	Limit int64 `protobuf:"varint,10,opt,name=limit,proto3" json:"limit,omitempty"`
	// pattern_matches_content is whether the pattern should be matched
	// against the content of files.
	PatternMatchesContent bool `protobuf:"varint,11,opt,name=pattern_matches_content,json=patternMatchesContent,proto3" json:"pattern_matches_content,omitempty"`
	// pattern_matches_content is whether a file whose path matches
	// pattern (but whose contents don't) should be considered a match.
	PatternMatchesPath bool `protobuf:"varint,12,opt,name=pattern_matches_path,json=patternMatchesPath,proto3" json:"pattern_matches_path,omitempty"`
	// comby_rule is a rule that constrains matching for structural search.
	// It only applies when IsStructuralPat is true.
	// As a temporary measure, the expression `where "backcompat" == "backcompat"`
	// acts as a flag to activate the old structural search path, which queries
	// zoekt for the file list in the frontend and passes it to searcher.
	CombyRule string `protobuf:"bytes,13,opt,name=comby_rule,json=combyRule,proto3" json:"comby_rule,omitempty"`
	// languages is the list of languages passed via the lang filters (e.g.,
	// "lang:c")
	Languages []string `protobuf:"bytes,14,rep,name=languages,proto3" json:"languages,omitempty"`
	// select is the value of the the select field in the query. It is not
	// necessary to use it since selection is done after the query completes, but
	// exposing it can enable optimizations.
	Select string `protobuf:"bytes,15,opt,name=select,proto3" json:"select,omitempty"`
}

func (x *PatternInfo) Reset() {
	*x = PatternInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_searcher_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PatternInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PatternInfo) ProtoMessage() {}

func (x *PatternInfo) ProtoReflect() protoreflect.Message {
	mi := &file_searcher_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PatternInfo.ProtoReflect.Descriptor instead.
func (*PatternInfo) Descriptor() ([]byte, []int) {
	return file_searcher_proto_rawDescGZIP(), []int{6}
}

func (x *PatternInfo) GetPattern() string {
	if x != nil {
		return x.Pattern
	}
	return ""
}

func (x *PatternInfo) GetIsNegated() bool {
	if x != nil {
		return x.IsNegated
	}
	return false
}

func (x *PatternInfo) GetIsRegexp() bool {
	if x != nil {
		return x.IsRegexp
	}
	return false
}

func (x *PatternInfo) GetIsStructural() bool {
	if x != nil {
		return x.IsStructural
	}
	return false
}

func (x *PatternInfo) GetIsCaseSensitive() bool {
	if x != nil {
		return x.IsCaseSensitive
	}
	return false
}

func (x *PatternInfo) GetExcludePattern() string {
	if x != nil {
		return x.ExcludePattern
	}
	return ""
}

func (x *PatternInfo) GetIncludePatterns() []string {
	if x != nil {
		return x.IncludePatterns
	}
	return nil
}

func (x *PatternInfo) GetPathPatternsAreCaseSensitive() bool {
	if x != nil {
		return x.PathPatternsAreCaseSensitive
	}
	return false
}

func (x *PatternInfo) GetLimit() int64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *PatternInfo) GetPatternMatchesContent() bool {
	if x != nil {
		return x.PatternMatchesContent
	}
	return false
}

func (x *PatternInfo) GetPatternMatchesPath() bool {
	if x != nil {
		return x.PatternMatchesPath
	}
	return false
}

func (x *PatternInfo) GetCombyRule() string {
	if x != nil {
		return x.CombyRule
	}
	return ""
}

func (x *PatternInfo) GetLanguages() []string {
	if x != nil {
		return x.Languages
	}
	return nil
}

func (x *PatternInfo) GetSelect() string {
	if x != nil {
		return x.Select
	}
	return ""
}

// Done is the final SearchResponse message sent in the stream
// of responses to Search.
type SearchResponse_Done struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LimitHit    bool `protobuf:"varint,1,opt,name=limit_hit,json=limitHit,proto3" json:"limit_hit,omitempty"`
	DeadlineHit bool `protobuf:"varint,2,opt,name=deadline_hit,json=deadlineHit,proto3" json:"deadline_hit,omitempty"`
}

func (x *SearchResponse_Done) Reset() {
	*x = SearchResponse_Done{}
	if protoimpl.UnsafeEnabled {
		mi := &file_searcher_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchResponse_Done) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchResponse_Done) ProtoMessage() {}

func (x *SearchResponse_Done) ProtoReflect() protoreflect.Message {
	mi := &file_searcher_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchResponse_Done.ProtoReflect.Descriptor instead.
func (*SearchResponse_Done) Descriptor() ([]byte, []int) {
	return file_searcher_proto_rawDescGZIP(), []int{1, 0}
}

func (x *SearchResponse_Done) GetLimitHit() bool {
	if x != nil {
		return x.LimitHit
	}
	return false
}

func (x *SearchResponse_Done) GetDeadlineHit() bool {
	if x != nil {
		return x.DeadlineHit
	}
	return false
}

var File_searcher_proto protoreflect.FileDescriptor

var file_searcher_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x0b, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x1a, 0x1e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xce, 0x02,
	0x0a, 0x0d, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x12, 0x0a, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x72,
	0x65, 0x70, 0x6f, 0x12, 0x17, 0x0a, 0x07, 0x72, 0x65, 0x70, 0x6f, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x72, 0x65, 0x70, 0x6f, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a,
	0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x5f, 0x6f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x4f, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x69,
	0x6e, 0x64, 0x65, 0x78, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x69, 0x6e,
	0x64, 0x65, 0x78, 0x65, 0x64, 0x12, 0x3b, 0x0a, 0x0c, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e,
	0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x73, 0x65,
	0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x74, 0x74, 0x65, 0x72,
	0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0b, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x49, 0x6e,
	0x66, 0x6f, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x75, 0x72, 0x6c, 0x12, 0x16, 0x0a, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x12, 0x3e, 0x0a, 0x0d,
	0x66, 0x65, 0x74, 0x63, 0x68, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0c,
	0x66, 0x65, 0x74, 0x63, 0x68, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x2a, 0x0a, 0x11,
	0x6e, 0x75, 0x6d, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x6c, 0x69, 0x6e, 0x65,
	0x73, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0f, 0x6e, 0x75, 0x6d, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x78, 0x74, 0x4c, 0x69, 0x6e, 0x65, 0x73, 0x4a, 0x04, 0x08, 0x09, 0x10, 0x0a, 0x22, 0xe3,
	0x01, 0x0a, 0x0e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x37, 0x0a, 0x0a, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x48, 0x00, 0x52,
	0x09, 0x66, 0x69, 0x6c, 0x65, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x12, 0x45, 0x0a, 0x0c, 0x64, 0x6f,
	0x6e, 0x65, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x20, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x44, 0x6f,
	0x6e, 0x65, 0x48, 0x00, 0x52, 0x0b, 0x64, 0x6f, 0x6e, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x1a, 0x46, 0x0a, 0x04, 0x44, 0x6f, 0x6e, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x6c, 0x69, 0x6d,
	0x69, 0x74, 0x5f, 0x68, 0x69, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x6c, 0x69,
	0x6d, 0x69, 0x74, 0x48, 0x69, 0x74, 0x12, 0x21, 0x0a, 0x0c, 0x64, 0x65, 0x61, 0x64, 0x6c, 0x69,
	0x6e, 0x65, 0x5f, 0x68, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x64, 0x65,
	0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x48, 0x69, 0x74, 0x42, 0x09, 0x0a, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x22, 0x7a, 0x0a, 0x09, 0x46, 0x69, 0x6c, 0x65, 0x4d, 0x61, 0x74, 0x63,
	0x68, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x3c, 0x0a, 0x0d, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x5f, 0x6d,
	0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x73,
	0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x75, 0x6e, 0x6b,
	0x4d, 0x61, 0x74, 0x63, 0x68, 0x52, 0x0c, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x4d, 0x61, 0x74, 0x63,
	0x68, 0x65, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x5f, 0x68, 0x69, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x48, 0x69, 0x74,
	0x22, 0x8e, 0x01, 0x0a, 0x0a, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x12,
	0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x3a, 0x0a, 0x0d, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x5f, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x15, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4c,
	0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x53, 0x74, 0x61, 0x72, 0x74, 0x12, 0x2a, 0x0a, 0x06, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x06, 0x72, 0x61, 0x6e, 0x67, 0x65,
	0x73, 0x22, 0x5d, 0x0a, 0x05, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x2b, 0x0a, 0x05, 0x73, 0x74,
	0x61, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x73, 0x65, 0x61, 0x72,
	0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x27, 0x0a, 0x03, 0x65, 0x6e, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x03, 0x65, 0x6e, 0x64,
	0x22, 0x4e, 0x0a, 0x08, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06,
	0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6f, 0x66,
	0x66, 0x73, 0x65, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6c, 0x69, 0x6e, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x04, 0x6c, 0x69, 0x6e, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x6f, 0x6c, 0x75,
	0x6d, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e,
	0x22, 0xab, 0x04, 0x0a, 0x0b, 0x50, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x49, 0x6e, 0x66, 0x6f,
	0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x69, 0x73,
	0x5f, 0x6e, 0x65, 0x67, 0x61, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09,
	0x69, 0x73, 0x4e, 0x65, 0x67, 0x61, 0x74, 0x65, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x69, 0x73, 0x5f,
	0x72, 0x65, 0x67, 0x65, 0x78, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x69, 0x73,
	0x52, 0x65, 0x67, 0x65, 0x78, 0x70, 0x12, 0x23, 0x0a, 0x0d, 0x69, 0x73, 0x5f, 0x73, 0x74, 0x72,
	0x75, 0x63, 0x74, 0x75, 0x72, 0x61, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x69,
	0x73, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x75, 0x72, 0x61, 0x6c, 0x12, 0x2a, 0x0a, 0x11, 0x69,
	0x73, 0x5f, 0x63, 0x61, 0x73, 0x65, 0x5f, 0x73, 0x65, 0x6e, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x69, 0x73, 0x43, 0x61, 0x73, 0x65, 0x53, 0x65,
	0x6e, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x65, 0x78, 0x63, 0x6c, 0x75,
	0x64, 0x65, 0x5f, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0e, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x50, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e,
	0x12, 0x29, 0x0a, 0x10, 0x69, 0x6e, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x5f, 0x70, 0x61, 0x74, 0x74,
	0x65, 0x72, 0x6e, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f, 0x69, 0x6e, 0x63, 0x6c,
	0x75, 0x64, 0x65, 0x50, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x73, 0x12, 0x46, 0x0a, 0x20, 0x70,
	0x61, 0x74, 0x68, 0x5f, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x73, 0x5f, 0x61, 0x72, 0x65,
	0x5f, 0x63, 0x61, 0x73, 0x65, 0x5f, 0x73, 0x65, 0x6e, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x18,
	0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x1c, 0x70, 0x61, 0x74, 0x68, 0x50, 0x61, 0x74, 0x74, 0x65,
	0x72, 0x6e, 0x73, 0x41, 0x72, 0x65, 0x43, 0x61, 0x73, 0x65, 0x53, 0x65, 0x6e, 0x73, 0x69, 0x74,
	0x69, 0x76, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x0a, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x36, 0x0a, 0x17, 0x70, 0x61, 0x74,
	0x74, 0x65, 0x72, 0x6e, 0x5f, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x08, 0x52, 0x15, 0x70, 0x61, 0x74, 0x74,
	0x65, 0x72, 0x6e, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x12, 0x30, 0x0a, 0x14, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x5f, 0x6d, 0x61, 0x74,
	0x63, 0x68, 0x65, 0x73, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x12, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x50,
	0x61, 0x74, 0x68, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x6f, 0x6d, 0x62, 0x79, 0x5f, 0x72, 0x75, 0x6c,
	0x65, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x6f, 0x6d, 0x62, 0x79, 0x52, 0x75,
	0x6c, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x73, 0x18,
	0x0e, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x73,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x4a, 0x04, 0x08, 0x05, 0x10, 0x06, 0x32, 0x5b,
	0x0a, 0x0f, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x48, 0x0a, 0x06, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x12, 0x1a, 0x2e, 0x73, 0x65,
	0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x03, 0x90, 0x02, 0x02, 0x30, 0x01, 0x42, 0x39, 0x5a, 0x37, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x67, 0x72, 0x61, 0x70,
	0x68, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x73, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_searcher_proto_rawDescOnce sync.Once
	file_searcher_proto_rawDescData = file_searcher_proto_rawDesc
)

func file_searcher_proto_rawDescGZIP() []byte {
	file_searcher_proto_rawDescOnce.Do(func() {
		file_searcher_proto_rawDescData = protoimpl.X.CompressGZIP(file_searcher_proto_rawDescData)
	})
	return file_searcher_proto_rawDescData
}

var file_searcher_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_searcher_proto_goTypes = []interface{}{
	(*SearchRequest)(nil),       // 0: searcher.v1.SearchRequest
	(*SearchResponse)(nil),      // 1: searcher.v1.SearchResponse
	(*FileMatch)(nil),           // 2: searcher.v1.FileMatch
	(*ChunkMatch)(nil),          // 3: searcher.v1.ChunkMatch
	(*Range)(nil),               // 4: searcher.v1.Range
	(*Location)(nil),            // 5: searcher.v1.Location
	(*PatternInfo)(nil),         // 6: searcher.v1.PatternInfo
	(*SearchResponse_Done)(nil), // 7: searcher.v1.SearchResponse.Done
	(*durationpb.Duration)(nil), // 8: google.protobuf.Duration
}
var file_searcher_proto_depIdxs = []int32{
	6,  // 0: searcher.v1.SearchRequest.pattern_info:type_name -> searcher.v1.PatternInfo
	8,  // 1: searcher.v1.SearchRequest.fetch_timeout:type_name -> google.protobuf.Duration
	2,  // 2: searcher.v1.SearchResponse.file_match:type_name -> searcher.v1.FileMatch
	7,  // 3: searcher.v1.SearchResponse.done_message:type_name -> searcher.v1.SearchResponse.Done
	3,  // 4: searcher.v1.FileMatch.chunk_matches:type_name -> searcher.v1.ChunkMatch
	5,  // 5: searcher.v1.ChunkMatch.content_start:type_name -> searcher.v1.Location
	4,  // 6: searcher.v1.ChunkMatch.ranges:type_name -> searcher.v1.Range
	5,  // 7: searcher.v1.Range.start:type_name -> searcher.v1.Location
	5,  // 8: searcher.v1.Range.end:type_name -> searcher.v1.Location
	0,  // 9: searcher.v1.SearcherService.Search:input_type -> searcher.v1.SearchRequest
	1,  // 10: searcher.v1.SearcherService.Search:output_type -> searcher.v1.SearchResponse
	10, // [10:11] is the sub-list for method output_type
	9,  // [9:10] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_searcher_proto_init() }
func file_searcher_proto_init() {
	if File_searcher_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_searcher_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_searcher_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_searcher_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileMatch); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_searcher_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChunkMatch); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_searcher_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Range); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_searcher_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Location); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_searcher_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PatternInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_searcher_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchResponse_Done); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_searcher_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*SearchResponse_FileMatch)(nil),
		(*SearchResponse_DoneMessage)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_searcher_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_searcher_proto_goTypes,
		DependencyIndexes: file_searcher_proto_depIdxs,
		MessageInfos:      file_searcher_proto_msgTypes,
	}.Build()
	File_searcher_proto = out.File
	file_searcher_proto_rawDesc = nil
	file_searcher_proto_goTypes = nil
	file_searcher_proto_depIdxs = nil
}
