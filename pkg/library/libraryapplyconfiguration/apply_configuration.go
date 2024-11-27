package libraryapplyconfiguration

import (
	"errors"
	"fmt"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryoutputresources"
	"path/filepath"
	"sort"
	"strings"

	"github.com/openshift/library-go/pkg/manifestclient"
	"k8s.io/apimachinery/pkg/util/sets"
)

// MutationActionReader provides access to serialized mutation requests
type MutationActionReader interface {
	ListActions() []manifestclient.Action
	RequestsForAction(action manifestclient.Action) []manifestclient.SerializedRequestish
	AllRequests() []manifestclient.SerializedRequestish
}

// SingleClusterDesiredMutationGetter provides access to mutations targeted at a single type of cluster
type SingleClusterDesiredMutationGetter interface {
	GetClusterType() ClusterType
	Requests() MutationActionReader
}

// AllDesiredMutationsGetter provides access to mutations targeted at all available types of clusters
type AllDesiredMutationsGetter interface {
	MutationsForClusterType(clusterType ClusterType) SingleClusterDesiredMutationGetter
}

type applyConfiguration struct {
	desiredMutationsByClusterType map[ClusterType]SingleClusterDesiredMutationGetter
}

var (
	_ AllDesiredMutationsGetter = &applyConfiguration{}
)

func UnspecifiedOutputResources(allDesiredMutationsGetter AllDesiredMutationsGetter, allAllowedOutputResources *libraryoutputresources.OutputResources) []manifestclient.SerializedRequestish {
	allMutationRequests := []manifestclient.SerializedRequestish{}
	for _, clusterType := range sets.List(AllClusterTypes) {
		desiredMutationsGetter := allDesiredMutationsGetter.MutationsForClusterType(clusterType)
		if desiredMutationsGetter != nil {
			allMutationRequests = append(allMutationRequests, desiredMutationsGetter.Requests().AllRequests()...)
		}
	}

	combinedList := &libraryoutputresources.ResourceList{}
	combinedList.ExactResources = append(combinedList.ExactResources, allAllowedOutputResources.ConfigurationResources.ExactResources...)
	combinedList.ExactResources = append(combinedList.ExactResources, allAllowedOutputResources.ManagementResources.ExactResources...)
	combinedList.ExactResources = append(combinedList.ExactResources, allAllowedOutputResources.UserWorkloadResources.ExactResources...)
	combinedList.GeneratedNameResources = append(combinedList.GeneratedNameResources, allAllowedOutputResources.ConfigurationResources.GeneratedNameResources...)
	combinedList.GeneratedNameResources = append(combinedList.GeneratedNameResources, allAllowedOutputResources.ManagementResources.GeneratedNameResources...)
	combinedList.GeneratedNameResources = append(combinedList.GeneratedNameResources, allAllowedOutputResources.UserWorkloadResources.GeneratedNameResources...)
	combinedList.EventingNamespaces = append(combinedList.EventingNamespaces, allAllowedOutputResources.ConfigurationResources.EventingNamespaces...)
	combinedList.EventingNamespaces = append(combinedList.EventingNamespaces, allAllowedOutputResources.ManagementResources.EventingNamespaces...)
	combinedList.EventingNamespaces = append(combinedList.EventingNamespaces, allAllowedOutputResources.UserWorkloadResources.EventingNamespaces...)
	filteredMutationRequests := FilterSerializedRequests(allMutationRequests, combinedList)

	return manifestclient.DifferenceOfSerializedRequests(allMutationRequests, filteredMutationRequests)
}

func ValidateAllDesiredMutationsGetter(allDesiredMutationsGetter AllDesiredMutationsGetter, allAllowedOutputResources *libraryoutputresources.OutputResources) error {
	errs := []error{}

	if allDesiredMutationsGetter == nil {
		return fmt.Errorf("applyConfiguration is required")
	}

	unspecifiedOutputResources := UnspecifiedOutputResources(allDesiredMutationsGetter, allAllowedOutputResources)
	if len(unspecifiedOutputResources) > 0 {
		unspecifiedOutputIdentifiers := []string{}
		for _, curr := range unspecifiedOutputResources {
			unspecifiedOutputIdentifiers = append(unspecifiedOutputIdentifiers, curr.GetSerializedRequest().StringID())
		}
		errs = append(errs, fmt.Errorf("%d output-resource were produced, but not present in the specified output: %v", len(unspecifiedOutputIdentifiers), strings.Join(unspecifiedOutputIdentifiers, ", ")))
	}

	return errors.Join(errs...)
}

func ValidateNoDuplicateMutations(allDesiredMutationsGetter AllDesiredMutationsGetter) error {
	var errs []error
	for _, clusterType := range sets.List(AllClusterTypes) {
		mutationsForClusterType := allDesiredMutationsGetter.MutationsForClusterType(clusterType)
		errs = append(errs, validateNoDuplicateMutationsForActions(mutationsForClusterType.Requests().RequestsForAction(manifestclient.ActionCreate), manifestclient.ActionCreate)...)
		// TODO: add Update, Deletion
	}
	return errors.Join(errs...)
}

func validateNoDuplicateMutationsForActions(requests []manifestclient.SerializedRequestish, action manifestclient.Action) []error {
	seenIDs := map[string]int{}
	for _, request := range requests {
		// TODO: skip if GenerateNam set ?
		id := request.GetSerializedRequest().StringID()
		idCounter := seenIDs[id]
		idCounter++
		seenIDs[id] = idCounter
	}

	sortedIDs := make([]string, 0, len(seenIDs))
	for id := range seenIDs {
		sortedIDs = append(sortedIDs, id)
	}
	sort.Strings(sortedIDs)

	var errs []error
	for _, id := range sortedIDs {
		idCounter := seenIDs[id]
		if idCounter > 1 {
			errs = append(errs, fmt.Errorf("detected multiple mutations for resource: %s, action: %s, mutations: %d", id, action, idCounter))
		}
	}
	return errs
}

func WriteApplyConfiguration(desiredApplyConfiguration AllDesiredMutationsGetter, outputDirectory string) error {
	errs := []error{}

	for _, clusterType := range sets.List(AllClusterTypes) {
		desiredMutations := desiredApplyConfiguration.MutationsForClusterType(clusterType)
		err := manifestclient.WriteMutationDirectory(filepath.Join(outputDirectory, string(clusterType)), desiredMutations.Requests().AllRequests()...)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed writing requests for %q: %w", clusterType, err))
		}
	}

	return errors.Join(errs...)
}

func (s *applyConfiguration) MutationsForClusterType(clusterType ClusterType) SingleClusterDesiredMutationGetter {
	return s.desiredMutationsByClusterType[clusterType]
}

type ClusterType string

var (
	ClusterTypeConfiguration ClusterType = "Configuration"
	ClusterTypeManagement    ClusterType = "Management"
	ClusterTypeUserWorkload  ClusterType = "UserWorkload"
	AllClusterTypes                      = sets.New(ClusterTypeConfiguration, ClusterTypeManagement, ClusterTypeUserWorkload)
)
