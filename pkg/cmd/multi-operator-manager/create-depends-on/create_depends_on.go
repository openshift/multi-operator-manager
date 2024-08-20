package create_depends_on

import (
	from_must_gather "github.com/deads2k/multi-operator-manager/pkg/cmd/multi-operator-manager/create-depends-on/from-must-gather"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func NewCreateDependsOnCommand(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "create-depends-on",
		SilenceErrors: true,
	}
	cmd.AddCommand(
		// TODO add a command that can take a kubeconfig and read the content from a live cluster
		from_must_gather.NewCreateDependsOnFromMustGatherCommand(streams),
		// TODO add command that can take failure output from a MOM apply-configuration call (includes resourceVersion) and walk through a resource-watch git repo to build the depends-on output for making an integration test.
	)
	return cmd
}
