package from_must_gather

import (
	"context"
	"fmt"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryinputresources"
	"os"
	"sigs.k8s.io/yaml"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type FromMustGatherFlags struct {
	MustGatherDirectory string

	// OutputDirectory is the directory to where output should be stored
	OutputDirectory string

	InputResourcesFile string
	OperatorBinary     string

	Streams genericiooptions.IOStreams
}

func NewCreateInputResourcesFromMustGatherFlags(streams genericiooptions.IOStreams) *FromMustGatherFlags {
	return &FromMustGatherFlags{
		Streams: streams,
	}
}

func NewCreateInputResourcesFromMustGatherCommand(streams genericiooptions.IOStreams) *cobra.Command {
	f := NewCreateInputResourcesFromMustGatherFlags(streams)

	cmd := &cobra.Command{
		Use:   "from-must-gather",
		Short: "Take a must-gather directory and operator (or depends-on output) and write the minimal output to disk.",

		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if err := f.Validate(); err != nil {
				return err
			}
			if err := f.Run(ctx); err != nil {
				return err
			}
			return nil
		},
	}

	f.BindFlags(cmd.Flags())

	return cmd
}

func (f *FromMustGatherFlags) BindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&f.MustGatherDirectory, "must-gather-dir", f.MustGatherDirectory, "The directory where must-gather output is located.")
	flags.StringVar(&f.OutputDirectory, "output-dir", f.OutputDirectory, "The directory where the output is stored.")
	flags.StringVar(&f.InputResourcesFile, "input-resources", f.InputResourcesFile, "The file where pertinent resources are stored.")
	flags.StringVar(&f.OperatorBinary, "operator-binary", f.OperatorBinary, "Path to the operator binary to call <operator-binary> depends-on.")
}

func (f *FromMustGatherFlags) Validate() error {
	if len(f.MustGatherDirectory) == 0 {
		return fmt.Errorf("--must-gather-dir is required")
	}
	if len(f.OutputDirectory) == 0 {
		return fmt.Errorf("--output-dir is required")
	}
	switch {
	case len(f.InputResourcesFile) == 0 && len(f.OperatorBinary) == 0:
		return fmt.Errorf("exactly one of --pertinent-resources and --operator-binary is required")
	case len(f.InputResourcesFile) == 0 && len(f.OperatorBinary) != 0:
		return fmt.Errorf("not yet wired through")
	case len(f.InputResourcesFile) != 0 && len(f.OperatorBinary) == 0:
	case len(f.InputResourcesFile) != 0 && len(f.OperatorBinary) != 0:
		return fmt.Errorf("exactly one of --pertinent-resources and --operator-binary is required")
	}
	return nil
}

func (f *FromMustGatherFlags) Run(ctx context.Context) error {
	pertinentResourcesBytes, err := os.ReadFile(f.InputResourcesFile)
	if err != nil {
		return fmt.Errorf("unable to read pertinent resources %q: %w", f.InputResourcesFile, err)
	}
	pertinentResources := &libraryinputresources.InputResources{}
	if err := yaml.Unmarshal(pertinentResourcesBytes, &pertinentResources); err != nil {
		return fmt.Errorf("unable to parse pertinent resources %q: %w", f.InputResourcesFile, err)
	}

	return libraryinputresources.WriteRequiredInputResourcesFromMustGather(ctx, pertinentResources, f.MustGatherDirectory, f.OutputDirectory)
}
