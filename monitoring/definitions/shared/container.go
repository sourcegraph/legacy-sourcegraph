package shared

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Container monitoring overviews - these provide short-term overviews of container
// behaviour for a service.
//
// These observables should only use cAdvisor metrics, and are thus only available on
// Kubernetes and docker-compose deployments.
//
// cAdvisor metrics reference: https://github.com/google/cadvisor/blob/master/docs/storage/prometheus.md#prometheus-container-metrics
const TitleContainerMonitoring = "Container monitoring (not available on server)"

var (
	ContainerRestarts sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:        "container_restarts",
			Description: "container restarts",
			// inspired by https://stackoverflow.com/a/63782891
			Query:          fmt.Sprintf(`(count by(name) (container_last_seen{%[1]s} unless (container_last_seen{%[1]s} offset 1m)) - 1) > 0`, CadvisorNameMatcher(containerName)),
			Warning:        monitoring.Alert().GreaterOrEqual(1),
			Panel:          monitoring.Panel().LegendFormat("{{name}}"),
			Owner:          owner,
			Interpretation: "This value is based on the number of times a container has not been seen in one minute.",
			PossibleSolutions: strings.Replace(`
				- **Kubernetes:**
					- Determine if the pod was OOM killed using 'kubectl describe pod {{CONTAINER_NAME}}' (look for 'OOMKilled: true') and, if so, consider increasing the memory limit in the relevant 'Deployment.yaml'.
					- Check the logs before the container restarted to see if there are 'panic:' messages or similar using 'kubectl logs -p {{CONTAINER_NAME}}'.
				- **Docker Compose:**
					- Determine if the pod was OOM killed using 'docker inspect -f \'{{json .State}}\' {{CONTAINER_NAME}}' (look for '"OOMKilled":true') and, if so, consider increasing the memory limit of the {{CONTAINER_NAME}} container in 'docker-compose.yml'.
					- Check the logs before the container restarted to see if there are 'panic:' messages or similar using 'docker logs {{CONTAINER_NAME}}' (note this will include logs from the previous and currently running container).
			`, "{{CONTAINER_NAME}}", containerName, -1),
		}
	}

	ContainerMemoryUsage sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:        "container_memory_usage",
			Description: "container memory usage by instance",
			Query:       fmt.Sprintf(`cadvisor_container_memory_usage_percentage_total{%s}`, CadvisorNameMatcher(containerName)),
			Warning:     monitoring.Alert().GreaterOrEqual(99),
			Panel:       monitoring.Panel().LegendFormat("{{name}}").Unit(monitoring.Percentage).Interval(100).Max(100).Min(0),
			Owner:       owner,
			PossibleSolutions: strings.Replace(`
			- **Kubernetes:** Consider increasing memory limit in relevant 'Deployment.yaml'.
			- **Docker Compose:** Consider increasing 'memory:' of {{CONTAINER_NAME}} container in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", containerName, -1),
		}
	}

	ContainerCPUUsage sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:        "container_cpu_usage",
			Description: "container cpu usage total (1m average) across all cores by instance",
			Query:       fmt.Sprintf(`cadvisor_container_cpu_usage_percentage_total{%s}`, CadvisorNameMatcher(containerName)),
			Warning:     monitoring.Alert().GreaterOrEqual(99),
			Panel:       monitoring.Panel().LegendFormat("{{name}}").Unit(monitoring.Percentage).Interval(100).Max(100).Min(0),
			Owner:       owner,
			PossibleSolutions: strings.Replace(`
			- **Kubernetes:** Consider increasing CPU limits in the the relevant 'Deployment.yaml'.
			- **Docker Compose:** Consider increasing 'cpus:' of the {{CONTAINER_NAME}} container in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", containerName, -1),
		}
	}

	// ContainerIOUsage monitors filesystem reads and writes, and is useful for services
	// that use disk.
	ContainerIOUsage sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:        "fs_io_operations",
			Description: "filesystem reads and writes rate by instance over 1h",
			Query:       fmt.Sprintf(`sum by(name) (rate(container_fs_reads_total{%[1]s}[1h]) + rate(container_fs_writes_total{%[1]s}[1h]))`, CadvisorNameMatcher(containerName)),
			NoAlert:     true,
			Panel:       monitoring.Panel().LegendFormat("{{name}}"),
			Owner:       monitoring.ObservableOwnerCloud,
			Interpretation: `
				This value indicates the number of filesystem read and write operations by containers of this service.
				When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.
			`,
		}
	}
)
