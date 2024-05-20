package internal

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
)

type PolicyType int

const (
	SyntacticIndexing PolicyType = 0
	PreciseIndexing   PolicyType = 1
)

// This iterator abstracts away the pagination logic for retrieving policies batches,
// propagating any errors
type PolicyIterator interface {
	// Iterate over all matching policies in batches. The `handle` function is NEVER
	// invoked with an empty policies list
	ForEachPoliciesBatch(ctx context.Context, handle func([]policiesshared.ConfigurationPolicy) error) error
}

type policyIterator struct {
	Service      policies.Service
	RepositoryID int
	PolicyType   PolicyType
	BatchSize    int
}

func (p policyIterator) ForEachPoliciesBatch(ctx context.Context, handle func([]policiesshared.ConfigurationPolicy) error) error {
	f := false
	t := true

	forSyntacticIndexing := &f
	forPreciseIndexing := &f

	if p.PolicyType == SyntacticIndexing {
		forSyntacticIndexing = &t
		forPreciseIndexing = nil
	} else {
		forPreciseIndexing = &t
		forSyntacticIndexing = nil
	}

	offset := 0

	for {

		options := policiesshared.GetConfigurationPoliciesOptions{
			RepositoryID:         p.RepositoryID,
			ForSyntacticIndexing: forSyntacticIndexing,
			ForPreciseIndexing:   forPreciseIndexing,
			Limit:                p.BatchSize,
			Offset:               offset,
		}

		policies, totalCount, err := p.Service.GetConfigurationPolicies(ctx, options)

		if err != nil {
			return err
		}

		if len(policies) == 0 {
			break
		}

		handlerError := handle(policies)

		if handlerError != nil {
			return handlerError // propagate error from the handler
		}

		offset = offset + len(policies)

		if offset >= totalCount {
			break
		}
	}

	return nil
}

var _ PolicyIterator = policyIterator{}

func NewPolicyIterator(service policies.Service, repositoryId int, policyType PolicyType, batchSize int) PolicyIterator {
	return policyIterator{
		Service:      service,
		RepositoryID: repositoryId,
		PolicyType:   policyType,
		BatchSize:    batchSize,
	}
}
