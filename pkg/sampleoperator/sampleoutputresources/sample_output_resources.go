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
				libraryoutputresources.ExactResource("config.openshift.io", "ingresses", "", "cluster"),
			}},
		ManagementResources: libraryoutputresources.ResourceList{
			ExactResources: []libraryoutputresources.ExactResourceID{
				libraryoutputresources.ExactResource("operator.openshift.io", "authentications", "", "cluster"),
			},
		},
		UserWorkloadResources: libraryoutputresources.ResourceList{
			ExactResources: []libraryoutputresources.ExactResourceID{
				libraryoutputresources.ExactResource("oauth.openshift.io", "oauthclients", "", "openshift-browser-client"),
			},
		},
	}, nil
}
