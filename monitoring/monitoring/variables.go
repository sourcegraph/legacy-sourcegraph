package monitoring

import (
	"strings"

	"github.com/grafana-tools/sdk"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/promql"
)

type ContainerVariableOptionType string

const (
	OptionTypeInterval = "interval"
)

type ContainerVariableOptions struct {
	Options []string
	// DefaultOption is the option that should be selected by default.
	DefaultOption string
	// Type of the options. You can usually leave this unset.
	Type ContainerVariableOptionType
}

type ContainerVariableOptionsQuery struct {
	Query string
	// ExampleOption is an example of a valid option for this variable that may be
	// generated by Query, and must be provided if using Query.
	ExampleOption string
}

// ContainerVariable describes a template variable that can be applied container dashboard
// for filtering purposes.
type ContainerVariable struct {
	// Name is the name of the variable to substitute the value for, e.g. "alert_level"
	// to replace "$alert_level" in queries
	Name string
	// Label is a human-readable name for the variable, e.g. "Alert level"
	Label string

	// OptionsQuery is the query to generate the possible values for this variable. Cannot
	// be used in conjunction with Options
	OptionsQuery ContainerVariableOptionsQuery
	// Options are the pre-defined possible values for this variable. Cannot be used in
	// conjunction with OptionsQuery
	Options ContainerVariableOptions

	// WildcardAllValue indicates to Grafana that is should NOT use OptionsQuery or Options to
	// generate a concatonated 'All' value for the variable, and use a '.*' wildcard
	// instead. Setting this to true primarily useful if you use Query and expect it to be
	// a large enough result set to cause issues when viewing the dashboard.
	//
	// We allow Grafana to generate a value by default because simply using '.*' wildcard
	// can pull in unintended metrics if adequate filtering is not performed on the query,
	// for example if multiple services export the same metric. If set to true, make sure
	// the queries that use this variable perform adequate filtering to avoid pulling in
	// unintended metrics.
	WildcardAllValue bool

	// Multi indicates whether or not to allow multi-selection for this variable filter
	Multi bool

	// RawTransform is can be used to extend ContainerVariable to modify underlying
	// Grafana variables specification.
	//
	// It is recommended to use or extend the standardized ContainerVariable options
	// instead.
	RawTransform func(*sdk.TemplateVar)
}

func (c *ContainerVariable) validate() error {
	if c.Name == "" {
		return errors.New("ContainerVariable.Name is required")
	}
	if c.Label == "" {
		return errors.New("ContainerVariable.Label is required")
	}
	if c.OptionsQuery.Query == "" && len(c.Options.Options) == 0 {
		return errors.New("one of ContainerVariable.OptionsQuery and ContainerVariable.Options must be set")
	}
	if c.OptionsQuery.Query != "" {
		if len(c.Options.Options) > 0 {
			return errors.New("ContainerVariable.OptionsQuery and ContainerVariable.Options cannot both be set")
		}
		if c.OptionsQuery.ExampleOption == "" {
			return errors.New("ContainerVariable.OptionsQuery.ExampleOption must be set")
		}
	}
	return nil
}

// toGrafanaTemplateVar generates the Grafana template variable configuration for this
// container variable.
func (c *ContainerVariable) toGrafanaTemplateVar() sdk.TemplateVar {
	variable := sdk.TemplateVar{
		Name:  c.Name,
		Label: c.Label,
		Multi: c.Multi,

		Datasource: StringPtr("Prometheus"),
		IncludeAll: true,

		// Apply the AllValue to a template variable by default
		Current: sdk.Current{Text: &sdk.StringSliceString{Value: []string{"all"}, Valid: true}, Value: "$__all"},
	}

	if c.WildcardAllValue {
		variable.AllValue = ".*"
	} else {
		// Rely on Grafana to create a union of only the values
		// generated by the specified query.
		//
		// See https://grafana.com/docs/grafana/latest/variables/formatting-multi-value-variables/#multi-value-variables-with-a-prometheus-or-influxdb-data-source
		// for more information.
		variable.AllValue = ""
	}

	switch {
	case c.OptionsQuery.Query != "":
		variable.Type = "query"
		variable.Query = c.OptionsQuery.Query
		variable.Refresh = sdk.BoolInt{
			Flag:  true,
			Value: Int64Ptr(2), // Refresh on time range change
		}
		variable.Sort = 3

	case len(c.Options.Options) > 0:
		// Set the type
		variable.Type = "custom"
		if c.Options.Type != "" {
			variable.Type = string(c.Options.Type)
		}
		// Generate our options
		variable.Query = strings.Join(c.Options.Options, ",")

		// On interval options, don't allow the selection of 'all' intervals, since
		// this is a one-of-many selection
		var hasAllOption bool
		if c.Options.Type != OptionTypeInterval {
			// Add the AllValue as a default, only selected if a default is not configured
			hasAllOption = true
			selected := c.Options.DefaultOption == ""
			variable.Options = append(variable.Options, sdk.Option{Text: "all", Value: "$__all", Selected: selected})
		}
		// Generate options
		for i, option := range c.Options.Options {
			// Whether this option should be selected
			var selected bool
			if c.Options.DefaultOption != "" {
				// If an default option is provided, select that
				selected = option == c.Options.DefaultOption
			} else if !hasAllOption {
				// Otherwise if there is no 'all' option generated, select the first
				selected = i == 0
			}

			variable.Options = append(variable.Options, sdk.Option{Text: option, Value: option, Selected: selected})
			if selected {
				// Also configure current
				variable.Current = sdk.Current{
					Text: &sdk.StringSliceString{
						Value: []string{option},
						Valid: true,
					},
					Value: option,
				}
			}
		}
	}

	if c.RawTransform != nil {
		c.RawTransform(&variable)
	}

	return variable
}

var numbers = regexp.MustCompile(`\d+`)

// getSentinelValue provides an easily distuingishable sentinel example value for this
// variable that allows a query with variables to be converted into a valid Prometheus
// query.
func (c *ContainerVariable) getSentinelValue() string {
	var example string
	switch {
	case len(c.Options.Options) > 0:
		example = c.Options.Options[0]
	case c.OptionsQuery.Query != "":
		example = c.OptionsQuery.ExampleOption
	default:
		return ""
	}
	// Scramble numerics - replace with a number that is very unlikely to conflict with
	// some other existing number in the query, this helps us distinguish what values
	// were replaced.
	return numbers.ReplaceAllString(example, "1234")
}

func newVariableApplier(vars []ContainerVariable) promql.VariableApplier {
	applier := promql.VariableApplier{}
	for _, v := range append(vars,
		// Make sure default Grafana variables are applied too.
		ContainerVariable{
			// https://grafana.com/docs/grafana/latest/datasources/prometheus/#using-__rate_interval
			Name: "__rate_interval",
			Options: ContainerVariableOptions{
				Options: []string{"123m"},
			},
		}) {
		applier[v.Name] = v.getSentinelValue()
	}
	return applier
}
