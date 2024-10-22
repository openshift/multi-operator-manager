package create_input_resources

import (
	from_must_gather "github.com/openshift/multi-operator-manager/pkg/cmd/multi-operator-manager/create-input-resources/from-must-gather"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func NewCreateInputResourcesCommand(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "create-input-resources",
		SilenceErrors: true,
	}
	cmd.AddCommand(
		// TODO add a command that can take a kubeconfig and read the content from a live cluster
		from_must_gather.NewCreateInputResourcesFromMustGatherCommand(streams),
		// TODO add command that can take failure output from a MOM apply-configuration call (includes resourceVersion) and walk through a resource-watch git repo to build the depends-on output for making an integration test.
	)
	return cmd
}
