package libraryapplyconfiguration

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	"github.com/openshift/library-go/pkg/manifestclient"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryoutputresources"
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
							Action: manifestclient.ActionApplyStatus,
							ResourceType: schema.GroupVersionResource{
								Group:    "",
								Version:  "v1",
								Resource: "secrets",
							},
							KindType: schema.GroupVersionKind{
								Group:   "",
								Version: "v1",
								Kind:    "Secret",
							},
							Namespace: "foo",
							Name:      "bar",
							Options:   nil,
							Body:      []byte(""),
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

func TestMutationsForControllerName(t *testing.T) {
	scenarios := []struct {
		name            string
		controllerName  string
		actualSecrets   []*corev1.Secret
		expectedSecrets []*corev1.Secret
	}{
		{
			name:           "controller name matches",
			controllerName: "fooController",
			actualSecrets: []*corev1.Secret{
				func() *corev1.Secret {
					secret := makeSecret("foo", "bar")
					secret.Annotations[manifestclient.ControllerNameAnnotation] = "fooController"
					return secret
				}()},
			expectedSecrets: []*corev1.Secret{
				func() *corev1.Secret {
					secret := makeSecret("foo", "bar")
					secret.Annotations[manifestclient.ControllerNameAnnotation] = "fooController"
					return secret
				}()},
		},
		{
			name:           "multiple secrets, controller name matches",
			controllerName: "fooController",
			actualSecrets: []*corev1.Secret{
				func() *corev1.Secret {
					secret := makeSecret("foo", "bar")
					secret.Annotations[manifestclient.ControllerNameAnnotation] = "fooController"
					return secret
				}(),
				makeSecret("foo1", "bar1"),
			},
			expectedSecrets: []*corev1.Secret{
				func() *corev1.Secret {
					secret := makeSecret("foo", "bar")
					secret.Annotations[manifestclient.ControllerNameAnnotation] = "fooController"
					return secret
				}()},
		},
		{
			name:           "controller name mismatch",
			controllerName: "barController",
			actualSecrets: []*corev1.Secret{
				func() *corev1.Secret {
					secret := makeSecret("foo", "bar")
					secret.Annotations[manifestclient.ControllerNameAnnotation] = "fooController"
					return secret
				}()},
		},
		{
			name: "empty result",
			actualSecrets: []*corev1.Secret{
				makeSecret("foo1", "bar1"),
			},
		},
		{
			name: "empty input",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			mutationTracker := manifestclient.NewAllActionsTracker[manifestclient.TrackedSerializedRequest]()
			for index, secret := range scenario.actualSecrets {
				index = index + 1
				secretBytes := toYAMLOrDie(secret)
				mutationTracker.AddRequest(manifestclient.TrackedSerializedRequest{
					RequestNumber: index,
					SerializedRequest: manifestclient.SerializedRequest{
						Action: manifestclient.ActionApplyStatus,
						ResourceType: schema.GroupVersionResource{
							Group:    "",
							Version:  "v1",
							Resource: "secrets",
						},
						KindType: schema.GroupVersionKind{
							Group:   "",
							Version: "v1",
							Kind:    "Secret",
						},
						Namespace: "foo",
						Name:      "bar",
						Options:   nil,
						Body:      secretBytes,
					},
				})
			}

			client := &clientBasedClusterApplyResult{
				clusterType:     ClusterTypeUserWorkload,
				mutationTracker: mutationTracker,
			}

			mutationsForControllerName, err := MutationsForControllerName(scenario.controllerName, client)
			if err != nil {
				t.Fatal(err)
			}
			var secretsWithControllerName []*corev1.Secret
			for _, mutationForControllerName := range mutationsForControllerName {
				secretsWithControllerName = append(secretsWithControllerName, secretFromYAMLOrDie(mutationForControllerName.GetSerializedRequest().Body))
			}

			if !cmp.Equal(scenario.expectedSecrets, secretsWithControllerName) {
				t.Fatalf("unexpected output, diff: %s", cmp.Diff(scenario.expectedSecrets, secretsWithControllerName))
			}
		})
	}
}

func makeSecret(namespace, name string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
	}
}

func toYAMLOrDie(obj runtime.Object) []byte {
	ret, err := yaml.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return ret
}

func secretFromYAMLOrDie(data []byte) *corev1.Secret {
	secret := &corev1.Secret{}
	if err := yaml.Unmarshal(data, secret); err != nil {
		panic(err)
	}
	return secret
}
