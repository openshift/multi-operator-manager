package sampleoutputresources

import (
	"context"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryoutputresources"
)

func SampleRunOutputResources(ctx context.Context) (*libraryoutputresources.OutputResources, error) {
	// TODO probably make this a yaml file directly?  I don't know, it's not too bad like this.
	return &libraryoutputresources.OutputResources{
		ConfigurationResources: libraryoutputresources.ResourceList{
			ExactResources: []libraryoutputresources.ExactResourceID{
				libraryoutputresources.ExactConfigResource("ingresses"),
			}},
		ManagementResources: libraryoutputresources.ResourceList{
			ExactResources: []libraryoutputresources.ExactResourceID{
				libraryoutputresources.ExactLowLevelOperator("authentications"),
			},
			EventingNamespaces: []string{
				"openshift-example-operator",
			},
		},
		UserWorkloadResources: libraryoutputresources.ResourceList{
			ExactResources: []libraryoutputresources.ExactResourceID{
				libraryoutputresources.ExactResource("oauth.openshift.io", "v1", "oauthclients", "", "openshift-browser-client"),
				libraryoutputresources.ExactConfigMap("openshift-authentication", "foo"),
			},
		},
	}, nil
}
