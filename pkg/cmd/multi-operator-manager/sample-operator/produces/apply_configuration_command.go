package applyconfiguration

import (
	"github.com/openshift/multi-operator-manager/pkg/library/libraryproduces"
	"github.com/openshift/multi-operator-manager/pkg/sampleoperator/sampleproduces"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func NewSampleOperatorProducesCommand(streams genericiooptions.IOStreams) *cobra.Command {
	return libraryproduces.NewProducesCommand(sampleproduces.SampleRunApplyProduces, streams)
}
