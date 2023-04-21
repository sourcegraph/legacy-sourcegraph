package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/streaming/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// NewCodeCompletionsHandler is an http handler which sends back code completion results
func NewCodeCompletionsHandler(logger log.Logger) http.Handler {
	return &codeCompletionHandler{logger: logger}
}

type codeCompletionHandler struct {
	logger log.Logger
}

func (h *codeCompletionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), maxRequestDuration)
	defer cancel()

	completionsConfig := conf.Get().Completions
	if completionsConfig == nil || !completionsConfig.Enabled {
		http.Error(w, "completions are not configured or disabled", http.StatusInternalServerError)
		return
	}

	if isEnabled := cody.IsCodyEnabled(ctx); !isEnabled {
		http.Error(w, "cody experimental feature flag is not enabled for current user", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
		return
	}

	var p types.CodeCompletionRequestParameters
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if honey.Enabled() {
		start := time.Now()
		defer func() {
			ev := honey.NewEvent("codeCompletions")
			ev.AddField("model", completionsConfig.CompletionModel)
			ev.AddField("actor", actor.FromContext(ctx).UIDString())
			ev.AddField("duration_sec", time.Since(start).Seconds())
			// This is the header which is useful for client IP on sourcegraph.com
			ev.AddField("connecting_ip", r.Header.Get("Cf-Connecting-Ip"))
			ev.AddField("ip_country", r.Header.Get("Cf-Ipcountry"))

			_ = ev.Send()
		}()
	}

	client := anthropic.NewAnthropicClient(httpcli.ExternalDoer, completionsConfig.AccessToken, completionsConfig.CompletionModel)
	completion, err := client.Complete(ctx, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	completionBytes, err := json.Marshal(completion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(completionBytes)
}
