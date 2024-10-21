package outputresources

import (
	"github.com/openshift/multi-operator-manager/pkg/library/libraryoutputresources"
	"github.com/openshift/multi-operator-manager/pkg/sampleoperator/sampleoutputresources"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func NewOutputResourcesCommand(streams genericiooptions.IOStreams) *cobra.Command {
	return libraryoutputresources.NewOutputResourcesCommand(sampleoutputresources.SampleRunOutputResources, streams)
}
