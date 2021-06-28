package gitlab

import (
	"context"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
)

// GetVersion retrieves the version of the Gitlab instance.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	time.Sleep(c.rateLimitMonitor.RecommendedWaitForBackgroundOp(1))

	var v struct {
		Version  string `json:"version"`
		Revision string `json:"revision"`
	}

	req, err := http.NewRequest("GET", "version", nil)
	if err != nil {
		return "", errors.Wrap(err, "creating version request")
	}

	_, _, err = c.do(ctx, req, &v)
	if err != nil {
		return "", errors.Wrap(err, "requesting version")
	}

	return v.Version, nil
}
