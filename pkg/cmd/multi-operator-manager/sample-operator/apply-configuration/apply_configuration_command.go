package applyconfiguration

import (
	"github.com/openshift/multi-operator-manager/pkg/library/libraryapplyconfiguration"
	"github.com/openshift/multi-operator-manager/pkg/sampleoperator/sampleapplyconfiguration"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func NewSampleOperatorApplyConfigurationCommand(streams genericiooptions.IOStreams) *cobra.Command {
	return libraryapplyconfiguration.NewApplyConfigurationCommand(sampleapplyconfiguration.SampleRunApplyConfiguration, streams)
}
