package httpcli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const keyPrefix = "outbound:"

func redisLoggerMiddleware() Middleware {
	creatorStackFrame, _ := getFrames(4).Next()
	return func(cli Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			resp, err := cli.Do(req)
			duration := time.Since(start)
			var requestBody []byte
			if req != nil {
				body, _ := req.GetBody()
				if body != nil {
					requestBody, _ = io.ReadAll(body)
				}
			}

			errorMessage := ""
			if err != nil {
				errorMessage = err.Error()
			}
			key := generateKey(time.Now())
			callerStackFrame, _ := getFrames(4).Next() // Caller of the caller of redisLoggerMiddleware
			logItem := types.OutboundRequestLogItem{
				Key:                key,
				StartedAt:          start,
				Method:             req.Method,
				URL:                req.URL.String(),
				RequestHeaders:     removeSensitiveHeaders(req.Header),
				RequestBody:        string(requestBody),
				StatusCode:         int32(resp.StatusCode),
				ResponseHeaders:    removeSensitiveHeaders(resp.Header),
				Duration:           duration.Seconds(),
				ErrorMessage:       errorMessage,
				CreationStackFrame: formatStackFrame(creatorStackFrame),
				CallStackFrame:     formatStackFrame(callerStackFrame),
			}

			logItemJson, jsonErr := json.Marshal(logItem)
			if jsonErr != nil {
				log.Error(jsonErr)
			}

			// Save new item
			redisCache.Set(key, logItemJson)

			deleteExcessItems()

			return resp, err
		})
	}
}

func deleteExcessItems() {
	deletionErr := redisCache.DeleteAllButLastN(keyPrefix, OutboundRequestLogLimit())
	if deletionErr != nil {
		log.Error(deletionErr)
	}
}

func generateKey(now time.Time) string {
	return fmt.Sprintf("%s%s", keyPrefix, now.UTC().Format("2006-01-02T15_04_05.999999999"))
}

// GetAllOutboundRequestLogItemsAfter returns all outbound request log items after the given key,
// in ascending order, trimmed to maximum {limit} items.
// The given "after" must contain `keyPrefix` as a prefix. Example: "outbound:2021-01-01T00_00_00.000000000".
func GetAllOutboundRequestLogItemsAfter(after *string, limit int) ([]*types.OutboundRequestLogItem, error) {
	if limit == 0 {
		return []*types.OutboundRequestLogItem{}, nil
	}

	rawItems, err := getAllValuesAfter(redisCache, keyPrefix, after, limit)
	if err != nil {
		return nil, err
	}

	items := make([]*types.OutboundRequestLogItem, 0, len(rawItems))
	for _, rawItem := range rawItems {
		var item types.OutboundRequestLogItem
		err = json.Unmarshal(rawItem, &item)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

// getAllValuesAfter returns all items after the given key, in ascending order, trimmed to maximum {limit} items.
// The given "after" must contain `keyPrefix` as a prefix. Example: "outbound:2021-01-01T00_00_00.000000000".
func getAllValuesAfter(redisCache *rcache.Cache, keyPrefix string, after *string, limit int) ([][]byte, error) {
	all, err := redisCache.ListKeysWithPrefix(keyPrefix)
	if err != nil {
		return nil, err
	}

	var keys []string
	if after != nil {
		for _, key := range all {
			if key > *after {
				keys = append(keys, key)
			}
		}
	} else {
		keys = all
	}

	// Sort ascending
	sort.Strings(keys)

	// Limit to N
	if len(keys) > limit {
		keys = keys[len(keys)-limit:]
	}

	return redisCache.GetMulti(keys...), nil
}

func removeSensitiveHeaders(headers http.Header) http.Header {
	var cleanHeaders = make(http.Header)
	for name, values := range headers {
		if IsRiskyHeader(name, values) {
			cleanHeaders[name] = []string{"REDACTED"}
		} else {
			cleanHeaders[name] = values
		}
	}
	return cleanHeaders
}

func formatStackFrame(frame runtime.Frame) string {
	packageTreeAndFunctionName := strings.Join(strings.Split(frame.Function, "/")[3:], "/")
	dotPieces := strings.Split(packageTreeAndFunctionName, ".")
	packageTree := dotPieces[0]
	functionName := dotPieces[len(dotPieces)-1]

	// Reconstruct the frame file path so that we don't include the local path on the machine that built this instance
	fileName := filepath.Join(packageTree, filepath.Base(frame.File))

	return fmt.Sprintf("%s:%d (Function: %s)", fileName, frame.Line, functionName)
}

const pcLen = 1024

func getFrames(skip int) *runtime.Frames {
	pc := make([]uintptr, pcLen)
	n := runtime.Callers(skip, pc)
	return runtime.CallersFrames(pc[:n])
}
