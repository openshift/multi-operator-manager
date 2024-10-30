package inputresources

import (
	"github.com/openshift/multi-operator-manager/pkg/library/libraryinputresources"
	"github.com/openshift/multi-operator-manager/pkg/sampleoperator/sampleinputresources"
	"github.com/openshift/multi-operator-manager/pkg/sampleoperator/sampleoutputresources"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func NewInputResourcesCommand(streams genericiooptions.IOStreams) *cobra.Command {
	return libraryinputresources.NewInputResourcesCommand(sampleinputresources.SampleRunInputResources, sampleoutputresources.SampleRunOutputResources, streams)
}
