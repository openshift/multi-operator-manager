package applyconfiguration

import (
	"fmt"
	"github.com/deads2k/multi-operator-manager/pkg/library/libraryapplyconfiguration"
	"reflect"

	"github.com/google/go-cmp/cmp"
)

func EquivalentApplyConfigurationResult(lhs, rhs ApplyConfigurationResult) []string {
	reasons := []string{}
	reasons = append(reasons, equivalentErrors("Error", lhs.Error(), rhs.Error())...)
	reasons = append(reasons, equivalentDesiredState("DesiredConfigurationCluster", lhs.DesiredConfigurationCluster, rhs.DesiredConfigurationCluster)...)
	reasons = append(reasons, equivalentDesiredState("DesiredManagementCluster", lhs.DesiredManagementCluster, rhs.DesiredManagementCluster)...)
	reasons = append(reasons, equivalentDesiredState("DesiredUserWorkloadCluster", lhs.DesiredUserWorkloadCluster, rhs.DesiredUserWorkloadCluster)...)

	return reasons
}

type desiredStateFunc func() (libraryapplyconfiguration.ClusterApplyResult, error)

func equivalentDesiredState(field string, lhs, rhs desiredStateFunc) []string {
	reasons := []string{}
	lhsDesiredState, lhsErr := lhs()
	rhsDesiredState, rhsErr := rhs()
	reasons = append(reasons, equivalentErrors(fmt.Sprintf("%s.Error", field), lhsErr, rhsErr)...)
	reasons = append(reasons, EquivalentClusterApplyResult(field, lhsDesiredState, rhsDesiredState)...)

	return reasons
}

func equivalentErrors(field string, lhs, rhs error) []string {
	reasons := []string{}
	switch {
	case lhs == nil && rhs == nil:
	case lhs == nil && rhs != nil:
		reasons = append(reasons, fmt.Sprintf("%v: lhs=nil, rhs=%v", field, rhs))
	case lhs != nil && rhs == nil:
		reasons = append(reasons, fmt.Sprintf("%v: lhs=%v, rhs=nil", field, lhs))
	case lhs.Error() != rhs.Error():
		reasons = append(reasons, fmt.Sprintf("%v: lhs=%v, rhs=%v", field, lhs, rhs))
	}

	return reasons
}

func EquivalentClusterApplyResult(field string, lhs, rhs libraryapplyconfiguration.ClusterApplyResult) []string {
	reasons := []string{}
	reasons = append(reasons, equivalentResourceFns(field+".Apply", lhs.ToApply, rhs.ToApply)...)
	reasons = append(reasons, equivalentResourceFns(field+".ToApplyStatus", lhs.ToApplyStatus, rhs.ToApplyStatus)...)
	reasons = append(reasons, equivalentResourceFns(field+".ToCreate", lhs.ToCreate, rhs.ToCreate)...)
	reasons = append(reasons, equivalentResourceFns(field+".ToUpdate", lhs.ToUpdate, rhs.ToUpdate)...)
	reasons = append(reasons, equivalentResourceFns(field+".ToUpdateStatus", lhs.ToUpdateStatus, rhs.ToUpdateStatus)...)
	reasons = append(reasons, equivalentResourceFns(field+".ToDelete", lhs.ToDelete, rhs.ToDelete)...)

	return reasons

}

type desiredResourcesFunc func() ([]libraryapplyconfiguration.Resource, error)

func equivalentResourceFns(field string, lhs, rhs desiredResourcesFunc) []string {
	reasons := []string{}
	lhsDesiredResources, lhsErr := lhs()
	rhsDesiredResources, rhsErr := rhs()
	reasons = append(reasons, equivalentErrors(field+".Error", lhsErr, rhsErr)...)
	reasons = append(reasons, equivalentResources(field, lhsDesiredResources, rhsDesiredResources)...)

	return reasons
}

func equivalentResources(field string, lhses, rhses []libraryapplyconfiguration.Resource) []string {
	reasons := []string{}

	for i := range lhses {
		lhs := lhses[i]
		rhs := findResource(rhses, lhs.Filename)

		if rhs == nil {
			reasons = append(reasons, fmt.Sprintf("%v[%d]: %q missing in rhs", field, i, lhs.Filename))
			continue
		}
		if !reflect.DeepEqual(lhs.Content, rhs.Content) {
			reasons = append(reasons, fmt.Sprintf("%v[%d]: does not match: %v", field, i, cmp.Diff(lhs.Content, rhs.Content)))
		}
	}

	for i := range rhses {
		rhs := rhses[i]
		lhs := findResource(lhses, rhs.Filename)

		if lhs == nil {
			reasons = append(reasons, fmt.Sprintf("%v[%d]: %q missing in lhs", field, i, rhs.Filename))
			continue
		}
	}

	return reasons
}

func findResource(in []libraryapplyconfiguration.Resource, filename string) *libraryapplyconfiguration.Resource {
	for _, curr := range in {
		if curr.Filename == filename {
			return &curr
		}
	}

	return nil
}
