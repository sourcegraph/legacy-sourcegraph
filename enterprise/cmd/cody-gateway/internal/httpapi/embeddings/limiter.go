package embeddings

import (
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func rateLimit(baseLogger log.Logger, eventLogger events.Logger, cache limiter.RedisStore, next http.Handler) http.Handler {
	baseLogger = baseLogger.Scoped("rateLimit", "rate limit handler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())
		logger := act.Logger(sgtrace.Logger(r.Context(), baseLogger))

		l, ok := act.EmbeddingsLimiter(cache)
		if !ok {
			response.JSONError(logger, w, http.StatusForbidden, errors.New("no access to embeddings"))
			return
		}

		commit, err := l.TryAcquire(r.Context())
		if err != nil {
			if loggerErr := eventLogger.LogEvent(
				r.Context(),
				events.Event{
					Name:       codygateway.EventNameRateLimited,
					Source:     act.Source.Name(),
					Identifier: act.ID,
					Metadata: map[string]any{
						"error": err.Error(),
						codygateway.CompletionsEventFeatureMetadataField: "embeddings",
					},
				},
			); loggerErr != nil {
				logger.Error("failed to log event", log.Error(loggerErr))
			}

			var rateLimitExceeded limiter.RateLimitExceededError
			if errors.As(err, &rateLimitExceeded) {
				rateLimitExceeded.WriteResponse(w)
				return
			}

			if errors.Is(err, limiter.NoAccessError{}) {
				response.JSONError(logger, w, http.StatusForbidden, err)
				return
			}

			response.JSONError(logger, w, http.StatusInternalServerError, err)
			return
		}

		responseRecorder := response.NewStatusHeaderRecorder(w)
		next.ServeHTTP(responseRecorder, r)

		// If response is healthy, consume the rate limit
		if responseRecorder.StatusCode >= 200 && responseRecorder.StatusCode < 300 {
			// TODO: We should commit the number of tokens consumed here, not simply increment by 1.
			if err := commit(); err != nil {
				logger.Error("failed to commit rate limit consumption", log.Error(err))
			}
		}
	})
}
