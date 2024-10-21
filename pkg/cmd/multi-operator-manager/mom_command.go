package multi_operator_manager

import (
	create_input_resources "github.com/openshift/multi-operator-manager/pkg/cmd/multi-operator-manager/create-input-resources"
	sample_operator "github.com/openshift/multi-operator-manager/pkg/cmd/multi-operator-manager/sample-operator"
	"github.com/openshift/multi-operator-manager/pkg/cmd/multi-operator-manager/test"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/logs"
	"k8s.io/component-base/version/verflag"
	"k8s.io/kubectl/pkg/util/templates"
)

func NewMultiOperatorManagerCommand(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use: "multi-operator-manager",
		Long: templates.LongDesc(`
		MultiOperatorManager

		This binary manages structured operator interactions with self-managed and externally managed topologies.
		It also provides structured test binaries to facilitate offline operator testing.
		`),
		SilenceErrors: true,
	}
	cmd.AddCommand(
		test.NewTestCommand(streams),
		sample_operator.NewSampleOperatorCommand(streams),
		create_input_resources.NewCreateInputResourcesCommand(streams),
	)

	verflag.AddFlags(cmd.Flags())
	globalflag.AddGlobalFlags(cmd.Flags(), cmd.Name(), logs.SkipLoggingConfigurationFlags())

	return cmd
}
