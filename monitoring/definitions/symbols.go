package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Symbols() *monitoring.Container {
	const (
		containerName = "symbols"
		primaryOwner  = monitoring.ObservableOwnerCodeIntel
	)

	return &monitoring.Container{
		Name:        "symbols",
		Title:       "Symbols",
		Description: "Handles symbol searches for unindexed branches.",
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						{
							Name:              "store_fetch_failures",
							Description:       "store fetch failures every 5m",
							Query:             `sum(increase(symbols_store_fetch_failed[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(5, nil),
							Panel:             monitoring.Panel().LegendFormat("failures"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "current_fetch_queue_size",
							Description:       "current fetch queue size",
							Query:             `sum(symbols_store_fetch_queue_size)`,
							Warning:           monitoring.Alert().GreaterOrEqual(25, nil),
							Panel:             monitoring.Panel().LegendFormat("size"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						shared.FrontendInternalAPIErrorResponses("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			shared.NewContainerMonitoringGroup(containerName, primaryOwner, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, primaryOwner, nil),
			shared.NewGolangMonitoringGroup(containerName, primaryOwner, nil),
			shared.NewKubernetesMonitoringGroup(containerName, primaryOwner, nil),
		},
	}
}
