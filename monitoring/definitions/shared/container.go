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

var (
	ContainerRestarts sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:        "container_restarts",
			Description: "container restarts every 5m by instance",
			Query:       fmt.Sprintf(`increase(cadvisor_container_restart_count{%s}[5m])`, CadvisorNameMatcher(containerName)),
			Warning:     monitoring.Alert().GreaterOrEqual(1),
			Panel:       monitoring.Panel().LegendFormat("{{name}}"),
			Owner:       owner,
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

	ContainerFsInodes sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:        "fs_inodes_used",
			Description: "fs inodes in use by instance",
			Query:       fmt.Sprintf(`sum by (name)(container_fs_inodes_total{%s})`, CadvisorNameMatcher(containerName)),
			Panel:       monitoring.Panel().LegendFormat("{{name}}"),
			NoAlert:     true,
			Owner:       owner,
			Interpretation: strings.Replace(`
				This value indicates the number of [filesystem inodes](https://en.wikipedia.org/wiki/Inode) held by containers of this service.
				When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself.

				If a high value or spikes here correlate with {{CONTAINER_NAME}} issues, the following might help:

				- **Increase available inodes**: Refer to your OS or cloud provider's documentation for how to increase inodes allowed on a machine.
				- **Kubernetes:** consider provisioning more machines for {{CONTAINER_NAME}} with less resources each.
			`, "{{CONTAINER_NAME}}", containerName, -1),
		}
	}
)
