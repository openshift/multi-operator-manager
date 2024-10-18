package libraryproduces

import (
	"context"
	"fmt"

	"k8s.io/cli-runtime/pkg/genericiooptions"
	"sigs.k8s.io/yaml"
)

type producesOptions struct {
	producesFn ProducesFunc

	streams genericiooptions.IOStreams
}

func newApplyConfigurationOptions(producesFn ProducesFunc, streams genericiooptions.IOStreams) *producesOptions {
	return &producesOptions{
		producesFn: producesFn,
		streams:    streams,
	}
}

func (o *producesOptions) Run(ctx context.Context) error {
	result, err := o.producesFn(ctx)
	if err != nil {
		return err
	}

	producesYaml, err := yaml.Marshal(result)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprint(o.streams.Out, string(producesYaml)); err != nil {
		return err
	}

	return nil
}
