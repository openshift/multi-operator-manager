package applyconfiguration

import (
	"context"
	"errors"
	"fmt"
	"github.com/openshift/multi-operator-manager/pkg/test/testapplyconfiguration"
	"io/fs"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type TestApplyConfigurationFlags struct {
	// TestDirectory is a directory that will be recursively walked to locate all directories containing a test
	// by finding directories which contain test.yaml
	// 1. test.yaml - serialized TestDescription
	// 2. input-dir - directory that will be provided to the apply-configuration command.
	// This allows for fairly arbitrary nesting strategies.
	TestDirectory string

	// OutputDirectory is the directory to where output should be stored
	OutputDirectory string

	PreservePolicy string

	Streams genericiooptions.IOStreams
}

func NewTestApplyConfigurationFlags(streams genericiooptions.IOStreams) *TestApplyConfigurationFlags {
	return &TestApplyConfigurationFlags{
		Streams: streams,
	}
}

func NewTestApplyConfigurationCommand(streams genericiooptions.IOStreams) *cobra.Command {
	f := NewTestApplyConfigurationFlags(streams)

	cmd := &cobra.Command{
		Use:   "apply-configuration",
		Short: "Test the apply-configuration command.",

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

func (f *TestApplyConfigurationFlags) BindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&f.TestDirectory, "test-dir", f.TestDirectory, "The directory where the tests are stored (recursive).")
	flags.StringVar(&f.OutputDirectory, "output-dir", f.OutputDirectory, "The directory where the output is stored.")
	flags.StringVar(&f.PreservePolicy, "preserve-policy", f.PreservePolicy, "")
}

func (f *TestApplyConfigurationFlags) Validate() error {
	if len(f.TestDirectory) == 0 {
		return fmt.Errorf("--test-dir is required")
	}
	if len(f.OutputDirectory) == 0 {
		return fmt.Errorf("--output-dir is required")
	}
	return nil
}

func (f *TestApplyConfigurationFlags) ToOptions(ctx context.Context) (*testapplyconfiguration.TestApplyConfigurationOptions, error) {
	tests := []testapplyconfiguration.TestOptions{}
	errs := []error{}
	err := filepath.WalkDir(f.TestDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			errs = append(errs, fmt.Errorf("%q: %w", path, err))
			return nil
		}
		if !d.IsDir() {
			return nil
		}

		currTest, _, err := testapplyconfiguration.ReadPotentialTestDir(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("%q: %w", path, err))
			return nil
		}
		if currTest == nil {
			return nil
		}

		outputDir := f.OutputDirectory
		if currentPathRelativeToInitialPath, err := filepath.Rel(f.TestDirectory, path); err == nil {
			outputDir = filepath.Join(outputDir, currentPathRelativeToInitialPath)
		}
		currTest.OutputDirectory = outputDir
		tests = append(tests, *currTest)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking: %w", err)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return &testapplyconfiguration.TestApplyConfigurationOptions{
		Tests:           tests,
		PreservePolicy:  f.PreservePolicy,
		OutputDirectory: f.OutputDirectory,
		Streams:         f.Streams,
	}, nil
}
