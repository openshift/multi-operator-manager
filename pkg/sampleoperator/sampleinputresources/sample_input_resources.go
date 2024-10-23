package sampleinputresources

import (
	"context"

	"github.com/openshift/multi-operator-manager/pkg/library/libraryinputresources"
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
			},
			ResourceReference: []libraryinputresources.ResourceReference{
				{
					ReferringResource: libraryinputresources.ExactConfigResource("ingresses"),
					Type:              "ImplicitNamespacedReference",
					ImplicitNamespacedReference: &libraryinputresources.ImplicitNamespacedReference{
						InputResourceTypeIdentifier: libraryinputresources.SecretIdentifierType(),
						Namespace:                   "openshift-config",
						NameJSONPath:                `.spec.componentRoutes[?(@.name == "my-route" && @.namespace == "openshift-authentication")].servingCertKeyPairSecret.name`,
					},
				},
			},
		},
		OperandResources: libraryinputresources.OperandResourceList{
			ConfigurationResources: libraryinputresources.ResourceList{},
			ManagementResources: libraryinputresources.ResourceList{
				ExactResources: []libraryinputresources.ExactResourceID{
					libraryinputresources.ExactDeployments("openshift-authentication", "oauth-server"),
				},
			},
			UserWorkloadResources: libraryinputresources.ResourceList{},
		},
	}, nil
}
