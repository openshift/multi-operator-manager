package sampleapplyconfiguration

import (
	"context"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryapplyconfiguration"
)

func SampleRunApplyConfiguration(ctx context.Context, input libraryapplyconfiguration.ApplyConfigurationInput) (libraryapplyconfiguration.AllDesiredMutationsGetter, error) {
	// TODO initialize dynamic clients, informers, operator clients, and kubeclients from the input to demonstrate.

	return libraryapplyconfiguration.NewApplyConfigurationFromClient(input.MutationTrackingClient.GetMutations()), nil
}
