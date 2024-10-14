package libraryapplyconfiguration

import (
	"context"
	"fmt"
	"time"

	"github.com/openshift/library-go/pkg/manifestclient"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/utils/clock"
	clocktesting "k8s.io/utils/clock/testing"
)

// ApplyConfigurationInput is provided to the ApplyConfigurationFunc
type ApplyConfigurationInput struct {
	// MutationTrackingClient is offered as an alternative to the inputDirectory to make it easier to provide mocks to code.
	// This forces all downstream code to rely on the client reading aspects and not grow an odd dependency to disk.
	MutationTrackingClient manifestclient.MutationTrackingClient

	// Now is the declared time that this function was called at.  It doesn't necessarily bear any relationship to
	// the actual time.  This is another aspect that makes unit and integration testing easier.
	Clock clock.Clock

	// Streams is for I/O.  The StdIn will usually be nil'd out.
	Streams genericiooptions.IOStreams
}

// ApplyConfigurationFunc is a function called for applying configuration.
type ApplyConfigurationFunc func(ctx context.Context, applyConfigurationInput ApplyConfigurationInput) (AllDesiredMutationsGetter, error)

func NewApplyConfigurationCommand(applyConfigurationFn ApplyConfigurationFunc, streams genericiooptions.IOStreams) *cobra.Command {
	return newSampleOperatorApplyConfigurationCommand(applyConfigurationFn, streams)
}

type sampleOperatorApplyConfigurationFlags struct {
	applyConfigurationFn ApplyConfigurationFunc

	// InputDirectory is a directory that contains the must-gather formatted inputs
	inputDirectory string

	// OutputDirectory is the directory to where output should be stored
	outputDirectory string

	streams genericiooptions.IOStreams
}

func newSampleOperatorApplyConfigurationFlags(streams genericiooptions.IOStreams) *sampleOperatorApplyConfigurationFlags {
	return &sampleOperatorApplyConfigurationFlags{
		streams: streams,
	}
}

func newSampleOperatorApplyConfigurationCommand(applyConfigurationFn ApplyConfigurationFunc, streams genericiooptions.IOStreams) *cobra.Command {
	f := newSampleOperatorApplyConfigurationFlags(streams)
	f.applyConfigurationFn = applyConfigurationFn

	cmd := &cobra.Command{
		Use:   "apply-configuration",
		Short: "Sample operator the apply-configuration command.",

		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(f.streams.ErrOut, "stderr output here\n")
			fmt.Fprintf(f.streams.Out, "stdout output here\n")
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

func (f *sampleOperatorApplyConfigurationFlags) BindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&f.inputDirectory, "input-dir", f.inputDirectory, "The directory where the resource input is stored.")
	flags.StringVar(&f.outputDirectory, "output-dir", f.outputDirectory, "The directory where the output is stored.")
}

func (f *sampleOperatorApplyConfigurationFlags) Validate() error {
	if len(f.inputDirectory) == 0 {
		return fmt.Errorf("--input-dir is required")
	}
	if len(f.outputDirectory) == 0 {
		return fmt.Errorf("--output-dir is required")
	}
	return nil
}

func (f *sampleOperatorApplyConfigurationFlags) ToOptions(ctx context.Context) (*applyConfigurationOptions, error) {
	momClient := manifestclient.NewHTTPClient(f.inputDirectory)
	input := ApplyConfigurationInput{
		MutationTrackingClient: momClient,
		Clock:                  clocktesting.NewFakeClock(time.Now()), // TODO fix to be an arg
		Streams:                f.streams,
	}

	return newApplyConfigurationOptions(
			f.applyConfigurationFn,
			input,
			f.outputDirectory,
		),
		nil
}
