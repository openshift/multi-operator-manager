package libraryapplyconfiguration

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type ApplyConfigurationFunc func(ctx context.Context, inputDirectory string, now time.Time, streams genericiooptions.IOStreams) (*ApplyConfiguration, error)

type ApplyConfigurationOptions struct {
	ApplyConfigurationFn ApplyConfigurationFunc

	InputDirectory string

	OutputDirectory string

	Now time.Time

	Streams genericiooptions.IOStreams
}

func NewApplyConfigurationOptions(
	applyConfigurationFn ApplyConfigurationFunc,
	inputDirectory string,
	outputDirectory string,
	now time.Time,
	streams genericiooptions.IOStreams) *ApplyConfigurationOptions {
	return &ApplyConfigurationOptions{
		ApplyConfigurationFn: applyConfigurationFn,
		InputDirectory:       inputDirectory,
		OutputDirectory:      outputDirectory,
		Now:                  now,
		Streams:              streams,
	}
}

func (o *ApplyConfigurationOptions) Run(ctx context.Context) error {
	if err := os.MkdirAll(o.OutputDirectory, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("unable to create output directory %q:%v", o.OutputDirectory, err)
	}

	result, err := o.ApplyConfigurationFn(ctx, o.InputDirectory, o.Now, o.Streams)
	if err != nil {
		return err
	}
	if err := result.Validate(); err != nil {
		return err
	}

	if err := WriteApplyConfiguration(result, o.OutputDirectory); err != nil {
		return err
	}

	return nil
}
