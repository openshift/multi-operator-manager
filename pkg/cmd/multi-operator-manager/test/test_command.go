package test

import (
	applyconfiguration "github.com/deads2k/multi-operator-manager/pkg/cmd/multi-operator-manager/test/apply-configuration"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func NewTestCommand(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "test",
		SilenceErrors: true,
	}
	cmd.AddCommand(
		applyconfiguration.NewTestApplyConfigurationCommand(streams),
	)
	return cmd
}
