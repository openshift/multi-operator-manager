package libraryapplyconfiguration

import (
	"fmt"
	"github.com/openshift/library-go/pkg/manifestclient"
	"k8s.io/apimachinery/pkg/util/sets"
)

func EquivalentApplyConfigurationResult(lhs, rhs ApplyConfigurationResult) []string {
	reasons := []string{}
	reasons = append(reasons, equivalentErrors("Error", lhs.Error(), rhs.Error())...)

	for _, clusterType := range sets.List(AllClusterTypes) {
		currLHS := lhs.MutationsForClusterType(clusterType)
		currRHS := rhs.MutationsForClusterType(clusterType)
		reasons = append(reasons, EquivalentClusterApplyResult(string(clusterType), currLHS, currRHS)...)
	}

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

func EquivalentClusterApplyResult(field string, lhs, rhs SingleClusterDesiredMutationGetter) []string {
	lhsRequests := lhs.Requests()
	rhsRequests := rhs.Requests()

	// TODO different method with prettier message
	equivalent := manifestclient.AreAllSerializedRequestsEquivalent(lhsRequests.AllRequests(), rhsRequests.AllRequests())
	if equivalent {
		return nil
	}

	return []string{
		"not equivalent",
	}
}
