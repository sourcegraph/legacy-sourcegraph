package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	enterprise_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

func main() {
	liblog := log.Init(log.Resource{
		Name:    env.MyName,
		Version: version.Version(),
	})
	defer liblog.Sync()

	logger := log.Scoped("worker", "worker enterprise edition")
	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	go enterprise_shared.SetAuthzProviders(logger, observationContext)

	if err := shared.Start(enterprise_shared.AdditionalJobs(observationContext), migrations.RegisterEnterpriseMigrators, logger, observationContext); err != nil {
		logger.Fatal(err.Error())
	}
}

func init() {
	oobmigration.ReturnEnterpriseMigrations = true
}
