package sampleoutputresources

import (
	"context"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryoutputresources"
)

func SampleRunOutputResources(ctx context.Context) (*libraryoutputresources.OutputResources, error) {
	// TODO probably make this a yaml file directly?  I don't know, it's not too bad like this.
	return &libraryoutputresources.OutputResources{
		ConfigurationResources: libraryoutputresources.ResourceList{},
		ManagementResources: libraryoutputresources.ResourceList{
			ExactResources: []libraryoutputresources.ExactResourceID{
				{
					OutputResourceTypeIdentifier: libraryoutputresources.OutputResourceTypeIdentifier{
						Group:    "config.openshift.io",
						Resource: "ingresses",
					},
					Namespace: "",
					Name:      "cluster",
				},
			},
		},
		UserWorkloadResources: libraryoutputresources.ResourceList{
			ExactResources: []libraryoutputresources.ExactResourceID{
				{
					OutputResourceTypeIdentifier: libraryoutputresources.OutputResourceTypeIdentifier{
						Group:    "oauth.openshift.io",
						Resource: "oauthclients",
					},
					Namespace: "",
					Name:      "openshift-browser-client",
				},
			},
		},
	}, nil
}
