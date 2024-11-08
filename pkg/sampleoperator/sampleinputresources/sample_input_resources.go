package sampleinputresources

import (
	"context"

	"github.com/openshift/multi-operator-manager/pkg/library/libraryinputresources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SampleRunInputResources(ctx context.Context) (*libraryinputresources.InputResources, error) {
	return &libraryinputresources.InputResources{
		ApplyConfigurationResources: libraryinputresources.ResourceList{
			ExactResources: []libraryinputresources.ExactResourceID{
				libraryinputresources.ExactSecret("openshift-authentication", "v4-0-config-system-ocp-branding-template"),
				libraryinputresources.ExactConfigResource("authentications"),
				libraryinputresources.ExactConfigResource("proxies"),
				libraryinputresources.ExactConfigResource("consoles"),
				libraryinputresources.ExactConfigResource("oauths"),
				libraryinputresources.ExactConfigMap("openshift-authentication", "fail-check"),
				libraryinputresources.ExactConfigMap("openshift-authentication", "foo"),
			},
			LabelSelectedResources: []libraryinputresources.LabelSelectedResource{
				{
					InputResourceTypeIdentifier: libraryinputresources.InputResourceTypeIdentifier{
						Group:    "",
						Version:  "v1",
						Resource: "configmaps",
					},
					Namespace: "openshift-oauth-apiserver",
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{"operator.openshift.io/controller-instance-name": "oauth-apiserver-RevisionController"},
					},
				},
				{
					InputResourceTypeIdentifier: libraryinputresources.InputResourceTypeIdentifier{
						Group:    "",
						Version:  "v1",
						Resource: "secrets",
					},
					Namespace: "openshift-oauth-apiserver",
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{"operator.openshift.io/controller-instance-name": "oauth-apiserver-RevisionController"},
					},
				},
			},
			ResourceReferences: []libraryinputresources.ResourceReference{
				{
					ReferringResource: libraryinputresources.ExactConfigResource("ingresses"),
					Type:              "ImplicitNamespacedReference",
					ImplicitNamespacedReference: &libraryinputresources.ImplicitNamespacedReference{
						InputResourceTypeIdentifier: libraryinputresources.SecretIdentifierType(),
						Namespace:                   "openshift-config",
						NameJSONPath:                `$.spec.componentRoutes[?(@.name == "my-route" && @.namespace == "openshift-authentication")].servingCertKeyPairSecret.name`,
					},
				},
			},
		},
		OperandResources: libraryinputresources.OperandResourceList{
			ConfigurationResources: libraryinputresources.ResourceList{},
			ManagementResources: libraryinputresources.ResourceList{
				ExactResources: []libraryinputresources.ExactResourceID{
					libraryinputresources.ExactDeployment("openshift-authentication", "oauth-server"),
				},
			},
			UserWorkloadResources: libraryinputresources.ResourceList{},
		},
	}, nil
}
