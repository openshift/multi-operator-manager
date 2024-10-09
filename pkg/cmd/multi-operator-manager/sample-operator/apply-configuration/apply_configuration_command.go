package applyconfiguration

import (
	"context"
	"fmt"
	"time"

	"github.com/openshift/multi-operator-manager/pkg/library/libraryapplyconfiguration"
	"github.com/openshift/multi-operator-manager/pkg/sampleoperator/sampleapplyconfiguration"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type SampleOperatorApplyConfigurationFlags struct {
	// InputDirectory is a directory that contains the must-gather formatted inputs
	InputDirectory string

	// OutputDirectory is the directory to where output should be stored
	OutputDirectory string

	Streams genericiooptions.IOStreams
}

func NewSampleOperatorApplyConfigurationFlags(streams genericiooptions.IOStreams) *SampleOperatorApplyConfigurationFlags {
	return &SampleOperatorApplyConfigurationFlags{
		Streams: streams,
	}
}

func NewSampleOperatorApplyConfigurationCommand(streams genericiooptions.IOStreams) *cobra.Command {
	f := NewSampleOperatorApplyConfigurationFlags(streams)

	cmd := &cobra.Command{
		Use:   "apply-configuration",
		Short: "Sample operator the apply-configuration command.",

		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(f.Streams.ErrOut, "stderr output here\n")
			fmt.Fprintf(f.Streams.Out, "stdout output here\n")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if err := f.Validate(); err != nil {
				return err
			}
			o, err := f.ToOptions(ctx)
			if err != nil {
				return err
			}
			if err := o.Run(ctx); err != nil {
				return err
			}
			return nil
		},
	}

	f.BindFlags(cmd.Flags())

	return cmd
}

func (f *SampleOperatorApplyConfigurationFlags) BindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&f.InputDirectory, "input-dir", f.InputDirectory, "The directory where the resource input is stored.")
	flags.StringVar(&f.OutputDirectory, "output-dir", f.OutputDirectory, "The directory where the output is stored.")
}

func (f *SampleOperatorApplyConfigurationFlags) Validate() error {
	if len(f.InputDirectory) == 0 {
		return fmt.Errorf("--input-dir is required")
	}
	if len(f.OutputDirectory) == 0 {
		return fmt.Errorf("--output-dir is required")
	}
	return nil
}

func (f *SampleOperatorApplyConfigurationFlags) ToOptions(ctx context.Context) (*libraryapplyconfiguration.ApplyConfigurationOptions, error) {
	return libraryapplyconfiguration.NewApplyConfigurationOptions(
			sampleapplyconfiguration.SampleRunApplyConfiguration,
			f.InputDirectory,
			f.OutputDirectory,
			time.Now(),
			f.Streams,
		),
		nil
}
