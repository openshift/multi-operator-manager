package sampleproduces

import (
	"context"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryproduces"
)

func SampleRunApplyProduces(ctx context.Context) (*libraryproduces.ProducedResources, error) {
	// TODO probably make this a yaml file directly?  I don't know, it's not too bad like this.
	return &libraryproduces.ProducedResources{
		ConfigurationServerResources: libraryproduces.ResourceList{},
		ManagementServerResources: libraryproduces.ResourceList{
			ExactResources: []libraryproduces.ExactResource{
				{
					ResourceTypeIdentifier: libraryproduces.ResourceTypeIdentifier{
						Group:    "config.openshift.io",
						Version:  "v1",
						Resource: "ingresses",
					},
					Namespace: "",
					Name:      "cluster",
				},
			},
		},
		GuestServerResources: libraryproduces.ResourceList{
			ExactResources: []libraryproduces.ExactResource{
				{
					ResourceTypeIdentifier: libraryproduces.ResourceTypeIdentifier{
						Group:    "oauth.openshift.io",
						Version:  "v1",
						Resource: "oauthclients",
					},
					Namespace: "",
					Name:      "openshift-browser-client",
				},
			},
		},
	}, nil
}
