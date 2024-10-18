package libraryproduces

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type ProducesFunc func(ctx context.Context) (*ProducedResources, error)

func NewProducesCommand(producesFn ProducesFunc, streams genericiooptions.IOStreams) *cobra.Command {
	return newProducesCommand(producesFn, streams)
}

type producesFlags struct {
	producesFn ProducesFunc

	streams genericiooptions.IOStreams
}

func newProducesFlags(streams genericiooptions.IOStreams) *producesFlags {
	return &producesFlags{
		streams: streams,
	}
}

func newProducesCommand(producesFn ProducesFunc, streams genericiooptions.IOStreams) *cobra.Command {
	f := newProducesFlags(streams)
	f.producesFn = producesFn

	cmd := &cobra.Command{
		Use:   "produces",
		Short: "Operator produces command.",

		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
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

func (f *producesFlags) BindFlags(flags *pflag.FlagSet) {
}

func (f *producesFlags) Validate() error {
	return nil
}

func (f *producesFlags) ToOptions(ctx context.Context) (*producesOptions, error) {
	return newApplyConfigurationOptions(
			f.producesFn,
			f.streams,
		),
		nil
}
