package libraryapplyconfiguration

import (
	"github.com/openshift/library-go/pkg/manifestclient"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryoutputresources"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestNewApplyConfigurationFromClient(t *testing.T) {
	type args struct {
		mutationTracker           func() *manifestclient.AllActionsTracker[manifestclient.TrackedSerializedRequest]
		allAllowedOutputResources *libraryoutputresources.OutputResources
	}
	tests := []struct {
		name   string
		args   args
		wanted func(t *testing.T, actual AllDesiredMutationsGetter)
	}{
		{
			name: "simple-assignment",
			args: args{
				mutationTracker: func() *manifestclient.AllActionsTracker[manifestclient.TrackedSerializedRequest] {
					ret := manifestclient.NewAllActionsTracker[manifestclient.TrackedSerializedRequest]()
					ret.AddRequest(manifestclient.TrackedSerializedRequest{
						RequestNumber: 1,
						SerializedRequest: manifestclient.SerializedRequest{
							ActionMetadata: manifestclient.ActionMetadata{
								Action: manifestclient.ActionApplyStatus,
								ResourceMetadata: manifestclient.ResourceMetadata{
									ResourceType: schema.GroupVersionResource{
										Group:    "",
										Version:  "v1",
										Resource: "secrets",
									},
									Namespace: "foo",
									Name:      "bar",
								},
							},
							KindType: schema.GroupVersionKind{
								Group:   "",
								Version: "v1",
								Kind:    "Secret",
							},
							Options: nil,
							Body:    []byte(""),
						},
					})
					return ret
				},
				allAllowedOutputResources: &libraryoutputresources.OutputResources{
					ConfigurationResources: libraryoutputresources.ResourceList{
						ExactResources: []libraryoutputresources.ExactResourceID{
							libraryoutputresources.ExactSecret("foo", "bar"),
						},
					},
					ManagementResources:   libraryoutputresources.ResourceList{},
					UserWorkloadResources: libraryoutputresources.ResourceList{},
				},
			},
			wanted: func(t *testing.T, actual AllDesiredMutationsGetter) {
				if a := actual.MutationsForClusterType(ClusterTypeManagement).Requests().AllRequests(); len(a) != 0 {
					t.Fatal(a)
				}
				if a := actual.MutationsForClusterType(ClusterTypeUserWorkload).Requests().AllRequests(); len(a) != 0 {
					t.Fatal(a)
				}
				configurationMutations := actual.MutationsForClusterType(ClusterTypeConfiguration)
				if a := configurationMutations.Requests().AllRequests(); len(a) != 1 {
					t.Fatal(a)
				}
				actualRequest := configurationMutations.Requests().AllRequests()[0]
				if e, a := "ApplyStatus-Secret.v1./bar[foo]", actualRequest.GetSerializedRequest().StringID(); e != a {
					t.Fatal(a)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intermediate := NewApplyConfigurationFromClient(tt.args.mutationTracker())

			got := FilterAllDesiredMutationsGetter(intermediate, tt.args.allAllowedOutputResources)
			tt.wanted(t, got)
		})
	}
}
