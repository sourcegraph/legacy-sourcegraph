package policies

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Configurations
	getConfigurationPolicies      *observation.Operation
	getConfigurationPoliciesByID  *observation.Operation
	createConfigurationPolicy     *observation.Operation
	updateConfigurationPolicy     *observation.Operation
	deleteConfigurationPolicyByID *observation.Operation

	// Retention Policy
	getRetentionPolicyOverview *observation.Operation

	// Repository
	getPreviewRepositoryFilter                  *observation.Operation
	getPreviewGitObjectFilter                   *observation.Operation
	selectPoliciesForRepositoryMembershipUpdate *observation.Operation
	updateReposMatchingPatterns                 *observation.Operation
}

var m = memo.NewMemoizedConstructorWithArg(func(r prometheus.Registerer) (*metrics.REDMetrics, error) {
	return metrics.NewREDMetrics(
		r,
		"codeintel_policies",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	), nil
})

func newOperations(observationContext *observation.Context) *operations {
	metrics, _ := m.Init(observationContext.Registerer)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.policies.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		// Configurations
		getConfigurationPolicies:      op("GetConfigurationPolicies"),
		getConfigurationPoliciesByID:  op("GetConfigurationPoliciesByID"),
		createConfigurationPolicy:     op("CreateConfigurationPolicy"),
		updateConfigurationPolicy:     op("UpdateConfigurationPolicy"),
		deleteConfigurationPolicyByID: op("DeleteConfigurationPolicyByID"),

		// Retention
		getRetentionPolicyOverview: op("GetRetentionPolicyOverview"),

		// Repository
		getPreviewRepositoryFilter:                  op("GetPreviewRepositoryFilter"),
		getPreviewGitObjectFilter:                   op("GetPreviewGitObjectFilter"),
		selectPoliciesForRepositoryMembershipUpdate: op("SelectPoliciesForRepositoryMembershipUpdate"),
		updateReposMatchingPatterns:                 op("UpdateReposMatchingPatterns"),
	}
}
