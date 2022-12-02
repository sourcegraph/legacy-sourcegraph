package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func OtelCollector() *monitoring.Dashboard {
	containerName := "otel-collector"

	return &monitoring.Dashboard{
		Name:        containerName,
		Title:       "Open Telemetry Collector",
		Description: "Metrics about the operation of the open telemetry collector.",
		Groups: []monitoring.Group{
			{
				Title:  "Receivers",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name:        "otel_span_receive_rate",
							Description: "spans received per second",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("receiver: {{receiver}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_receiver_accepted_spans{receiver=~\"^.*\"}[1m])) by (receiver)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans accepted by the configured reveiver
								
								A Trace is a collection of spans and a span represents a unit of work or operation. Spans are the building blocks of Traces.
								The spans have only been accepted by the receiver, which means they still have to move through the configured pipeline to be exported.
								For more information on tracing and configuration of a OpenTelemetry receiver see https://opentelemetry.io/docs/collector/configuration/#receivers.
								
								See the Exporters section see spans that have made it through the pipeline and are exported.
								
								Depending the configured processors, received spans might be dropped and not exported. For more information on configuring processors see
								https://opentelemetry.io/docs/collector/configuration/#processors.`,
						},
						{
							Name:        "otel_span_refused",
							Description: "spans that the receiver refused",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("receiver: {{receiver}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_receiver_refused_spans{receiver=~\"^.*.*\"}[1m])) by (receiver)",
							NoAlert:     true,
							Interpretation: `
								Shows the amount of spans that have been refused by a receiver.
								
								A Trace is a collection of spans. A Span represents a unit of work or operation. Spans are the building blocks of Traces.
							
 								Spans can be rejected either due to a misconfigured receiver or receiving spans in the wrong format. The log of the collector will have more information on why a span was rejected.
								For more information on tracing and configuration of a OpenTelemetry receiver see https://opentelemetry.io/docs/collector/configuration/#receivers.`,
						},
					},
				},
			},
			{
				Title:  "Exporters",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name:        "otel_span_export_rate",
							Description: "spans exported per second",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("exporter: {{exporter}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_exporter_sent_spans{exporter=~\"^.*\"}[1m])) by (exporter)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans being sent by the exporter
 								
								A Trace is a collection of spans. A Span represents a unit of work or operation. Spans are the building blocks of Traces.
								The rate of spans here indicates spans that have made it through the configured pipeline and have been sent to the configured export destination.
								
								For more information on configuring a exporter for the OpenTelemetry collector see https://opentelemetry.io/docs/collector/configuration/#exporters.`,
						},
						{
							Name:        "otel_span_failed_send_size",
							Description: "spans that the exporter failed to send",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("exporter: {{exporter}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_exporter_send_failed_spans{exporter=~\"^.*\"}[1m])) by (exporter)",
							NoAlert:     true,
							Interpretation: `
								Shows the rate of spans failed to be sent by the configured reveiver. A number higher than 0 for a long period can indicate a problem with the exporter configuration or with the service that is being exported too
								
								For more information on configuring a exporter for the OpenTelemetry collector see https://opentelemetry.io/docs/collector/configuration/#exporters.`,
						},
						{
							Name:        "otel_span_queue_size",
							Description: "spans pending to be sent",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("exporter: {{exporter}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_exporter_queue_size{exporter=~\"^.*\"}[1m])) by (exporter)",
							NoAlert:     true,
							Interpretation: `
								Indicates the amount of spans that are in the queue to be sent (exported). A high queue count might indicate a high volume of spans or a problem with the receiving service to which spans are being exported too
								
								For more information on configuring a exporter for the OpenTelemetry collector see https://opentelemetry.io/docs/collector/configuration/#exporters.`,
						},
						{
							Name:        "otel_span_queue_capacity",
							Description: "spans max items that can be pending to be sent",
							Panel:       monitoring.Panel().Unit(monitoring.Number).LegendFormat("exporter: {{exporter}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_exporter_queue_capacity{exporter=~\"^.*\"}[1m])) by (exporter)",
							NoAlert:     true,
							Interpretation: `
								Indicates the amount of spans that are in the queue to be sent (exported). A high queue count might indicate a high volume of spans or a problem with the receiving service
								
								For more information on configuring a exporter for the OpenTelemetry collector see https://opentelemetry.io/docs/collector/configuration/#exporters.`,
						},
					},
				},
			},
			{
				Title:  "Collector resource usage",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name:        "otel_cpu_usage",
							Description: "cpu usage of the collector",
							Panel:       monitoring.Panel().Unit(monitoring.Seconds).LegendFormat("{{job}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_process_cpu_seconds{job=~\"^.*\"}[1m])) by (job)",
							NoAlert:     true,
							Interpretation: `
								Shows the cpu usage of the OpenTelemetry collector`,
						},
						{
							Name:        "otel_memory_resident_set_size",
							Description: "memory allocated to the otel collector",
							Panel:       monitoring.Panel().Unit(monitoring.Bytes).LegendFormat("{{job}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_process_memory_rss{job=~\"^.*\"}[1m])) by (job)",
							NoAlert:     true,
							Interpretation: `
								Shows the memory Resident Set Size (RSS) allocated to the OpenTelemetry collector`,
						},
						{
							Name:        "otel_memory_usage",
							Description: "memory used by the collector",
							Panel:       monitoring.Panel().Unit(monitoring.Bytes).LegendFormat("{{job}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Query:       "sum(rate(otelcol_process_runtime_total_alloc_bytes{job=~\"^.*\"}[1m])) by (job)",
							NoAlert:     true,
							Interpretation: `
								Shows how much memory is being used by the otel collector.
								
								* High memory usage might indicate thad the configured pipeline is keeping a lot of spans in memory for processing
								* Spans failing to be sent and the exporter is configured to retry
								* A high bacth count by using a batch processor
								
								For more information on configuring processors for the OpenTelemetry collector see https://opentelemetry.io/docs/collector/configuration/#processors.`,
						},
					},
				},
			},
			shared.NewContainerMonitoringGroup("otel-collector", monitoring.ObservableOwnerDevOps, nil),
			shared.NewKubernetesMonitoringGroup("otel-collector", monitoring.ObservableOwnerDevOps, nil),
		},
	}
}
