package sample_operator

import (
	applyconfiguration "github.com/openshift/multi-operator-manager/pkg/cmd/multi-operator-manager/sample-operator/apply-configuration"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func NewSampleOperatorCommand(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "sample-operator",
		SilenceErrors: true,
	}
	cmd.AddCommand(
		applyconfiguration.NewSampleOperatorApplyConfigurationCommand(streams),
	)

	return cmd
}
