package httpapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/gorilla/mux"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/usagestats"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/eventlogger"
)

var telemetryHandler http.Handler

func init() {
	if envvar.SourcegraphDotComMode() {
		telemetryHandler = &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "https"
				req.URL.Host = "sourcegraph-logging.telligentdata.com"
				req.Host = "sourcegraph-logging.telligentdata.com"
				req.URL.Path = "/" + mux.Vars(req)["TelemetryPath"]
			},
			ErrorLog: log.New(env.DebugOut, "telemetry proxy: ", log.LstdFlags),
		}
	} else {
		telemetryHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req eventlogger.TelemetryRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				log15.Error("Decode", "error", err)
			}
			fmt.Println("===== Telemetry received by server: " + req.EventLabel + ", " + strconv.Itoa(int(req.UserID)))
			if req.UserID != 0 && req.EventLabel == "SavedSearchEmailNotificationSent" {
				usagestats.LogActivity(true, req.UserID, "", "STAGEVERIFY")
			}

			fmt.Fprintln(w, "event-level telemetry is disabled")
			w.WriteHeader(http.StatusNoContent)
		})
	}
}
