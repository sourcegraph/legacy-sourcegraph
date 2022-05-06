package shared

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

func Test_manualPurgeHandler(t *testing.T) {
	db := database.NewMockDB()
	handler := manualPurgeHandler(db)

	for _, tt := range []struct {
		name     string
		url      string
		wantCode int
		wantBody string
	}{
		{
			name:     "missing limit",
			url:      "https://example.com/manual_purge",
			wantCode: http.StatusBadRequest,
			wantBody: `invalid limit: strconv.Atoi: parsing "": invalid syntax
`,
		},
		{
			name:     "zero limit",
			url:      "https://example.com/manual_purge?limit=0",
			wantCode: http.StatusBadRequest,
			wantBody: `limit must be greater than 0
`,
		},
		{
			name:     "missing perSecond, default 1.0",
			url:      "https://example.com/manual_purge?limit=100",
			wantCode: http.StatusOK,
			wantBody: `manual purge started with limit of 100 and rate of 1.000000`,
		},
		{
			name:     "invalid perSecond",
			url:      "https://example.com/manual_purge?limit=100&perSecond=0",
			wantCode: http.StatusBadRequest,
			wantBody: `invalid per second rate limit. Must be > 0, got 0.000000
`,
		},
		{
			name:     "valid perSecond",
			url:      "https://example.com/manual_purge?limit=100&perSecond=2.0",
			wantCode: http.StatusOK,
			wantBody: `manual purge started with limit of 100 and rate of 2.000000`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}
			handler(rr, req)
			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, rr.Body.String())
		})
	}
}
