package testapplyconfiguration

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"time"

	"github.com/openshift/library-go/test/library/junitapi"
	"github.com/openshift/multi-operator-manager/pkg/applyconfiguration"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type TestApplyConfigurationOptions struct {
	// TestDirectory is a directory that will be recursively walked to locate all directories containing a test
	// by finding directories which contain test.yaml
	// 1. test.yaml - serialized TestDescription
	// 2. input-dir - directory that will be provided to the apply-configuration command.
	// 3. output-dir - directory that is expected from apply-configuration command.
	// This allows for fairly arbitrary nesting strategies.
	Tests []TestOptions

	// JunitSuiteName allows naming the junit suite for convenience.  Sometimes we run the same tests before/after a change
	// or in slightly different circumstances. This lets us accomodate that.
	JunitSuiteName string

	OutputDirectory string

	PreservePolicy string

	Streams genericiooptions.IOStreams
}

type TestOptions struct {
	Description TestDescription
	// TestDirectory is the directory containing the test to run. The directory containing test.yaml and input-dir
	TestDirectory string
	// OutputDirectory is the directory where the output should be
	OutputDirectory string
}

// now is available for unit tests
var now = time.Now

func (o *TestApplyConfigurationOptions) Run(ctx context.Context) error {
	junitFile := filepath.Join(o.OutputDirectory, "junit.xml")

	junit := &junitapi.JUnitTestSuite{
		XMLName: xml.Name{},
		Name:    o.JunitSuiteName,
		// TODO information if we want it.
		//Properties: []*junitapi.TestSuiteProperty{
		//	{
		//		Name:  "TestVersion",
		//		Value: version.Get().String(),
		//	},
		//},
	}
	defer func() {
		junitBytes, err := xml.MarshalIndent(junit, "", "    ")
		if err != nil {
			utilruntime.HandleError(err)
			return
		}
		if err := os.WriteFile(junitFile, junitBytes, 0644); err != nil {
			utilruntime.HandleError(err)
			return
		}
	}()

	if err := os.MkdirAll(o.OutputDirectory, 0755); err != nil && !os.IsExist(err) {
		retErr := fmt.Errorf("unable to create output directory %q:%v", o.OutputDirectory, err)
		junit.TestCases = append(junit.TestCases, &junitapi.JUnitTestCase{
			Name: "ensure output directory",
			FailureOutput: &junitapi.FailureOutput{
				Message: retErr.Error(),
				Output:  retErr.Error(),
			},
		})
		return retErr
	} else {
		junit.TestCases = append(junit.TestCases, &junitapi.JUnitTestCase{
			Name: "ensure output directory",
		})
	}

	failedTests := sets.Set[string]{}
	for _, test := range o.Tests {
		if ctx.Err() != nil {
			// break the loop and report as much as we can.
			break
		}
		currJunit := test.runTest(ctx)
		junit.TestCases = append(junit.TestCases, currJunit)
		if currJunit.FailureOutput != nil {
			failedTests.Insert(currJunit.Name)

		} else {
			// if we succeeded, we might need to cleanup the output
			if len(o.PreservePolicy) == 0 {
				if err := os.RemoveAll(test.OutputDirectory); err != nil {
					utilruntime.HandleError(fmt.Errorf("unable to cleanup for %q: %w", test.Description.TestName, err))
				}
			}
		}
	}

	junitBytes, err := xml.MarshalIndent(junit, "", "    ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(junitFile, junitBytes, 0644); err != nil {
		return err
	}

	if len(failedTests) > 0 {
		fmt.Fprintf(o.Streams.ErrOut, "%d tests failed:\n%v\n", len(failedTests), strings.Join(sets.List(failedTests), "\n"))
		return fmt.Errorf("%d tests failed", len(failedTests))
	}
	if ctx.Err() != nil {
		fmt.Fprintf(o.Streams.ErrOut, "failing due to cancellation (possibly timeout): %v", ctx.Err())
		return ctx.Err()
	}

	return nil
}

func (test *TestOptions) runTest(ctx context.Context) *junitapi.JUnitTestCase {
	junitTestName := fmt.Sprintf("%v [Binary:%q] [Directory:%q]", test.Description.TestName, test.Description.BinaryName, test.TestDirectory)
	currJunit := &junitapi.JUnitTestCase{
		Name: junitTestName,
	}
	startTime := now()

	if err := os.MkdirAll(test.OutputDirectory, 0755); err != nil && !os.IsExist(err) {
		currJunit.FailureOutput = &junitapi.FailureOutput{
			Message: fmt.Sprintf("unable to create output directory %q:\n%v\n", test.OutputDirectory, err),
			Output:  fmt.Sprintf("unable to create output directory %q:\n%v\n", test.OutputDirectory, err),
		}
		return currJunit
	}

	inputDir := filepath.Join(test.TestDirectory, "input-dir")
	actualResult, err := applyconfiguration.ApplyConfiguration(ctx, test.Description.BinaryName, inputDir, test.OutputDirectory)
	endTime := now()
	currJunit.Duration = endTime.Sub(startTime).Round(1 * time.Second).Seconds()

	switch {
	case err == nil && actualResult != nil:
		// this was successful
	case err == nil && actualResult == nil:
		currJunit.FailureOutput = &junitapi.FailureOutput{
			Message: "No result or error from apply-configuration",
			Output:  "No result or error from apply-configuration",
		}
		return currJunit

	case err != nil && actualResult != nil:
		currJunit.SystemOut = actualResult.Stdout()
		currJunit.SystemErr = actualResult.Stderr()
		fallthrough
	case err != nil && actualResult == nil:
		currJunit.FailureOutput = &junitapi.FailureOutput{
			Message: fmt.Sprintf("%v\n%v", err, currJunit.SystemErr),
			Output:  fmt.Sprintf("ERROR:%v\n\nSTDERR:\n%s\n\nSTDOUT:\n:%s\n", err, currJunit.SystemErr, currJunit.SystemOut),
		}
		return currJunit
	}

	expectedOutputDir := filepath.Join(test.TestDirectory, "expected-output")
	expectedResult, err := applyconfiguration.NewApplyConfigurationResult(expectedOutputDir, nil)
	if err != nil {
		currJunit.FailureOutput = &junitapi.FailureOutput{
			Message: fmt.Sprintf("failed to read expected output:\n%v\n", err),
			Output:  fmt.Sprintf("failed to read expected output:\n%v\n", err),
		}
		return currJunit
	}
	differences := applyconfiguration.EquivalentApplyConfigurationResult(expectedResult, actualResult)
	if len(differences) > 0 {
		currJunit.FailureOutput = &junitapi.FailureOutput{
			Message: fmt.Sprintf("expected results mismatch %d times with actual results", len(differences)),
			Output:  strings.Join(differences, "\n"),
		}
		return currJunit
	}

	return currJunit
}

var (
	requiredTestContent = sets.New("test.yaml", "input-dir", "expected-output")
)

func ReadPotentialTestDir(path string) (*TestOptions, bool, error) {
	actualContent, err := os.ReadDir(path)
	if err != nil {
		return nil, false, err
	}

	hasTestYaml := false
	for _, curr := range actualContent {
		if curr.Name() == "test.yaml" {
			hasTestYaml = true
		}
	}
	if !hasTestYaml {
		return nil, false, nil
	}

	missingContent := sets.Set[string]{}
	for _, requiredName := range requiredTestContent.UnsortedList() {
		found := false
		for _, curr := range actualContent {
			if curr.Name() == requiredName {
				found = true
			}
		}
		if !found {
			missingContent.Insert(requiredName)
		}
	}
	if len(missingContent) > 0 {
		return nil, true, fmt.Errorf("%q is missing: %v", path, sets.List(missingContent))
	}

	testYaml := filepath.Join(path, "test.yaml")
	testYamlBytes, err := os.ReadFile(testYaml)
	if err != nil {
		return nil, true, fmt.Errorf("unable to read test.yaml: %w", err)
	}
	testDescription := &TestDescription{}
	if err := yaml.Unmarshal(testYamlBytes, testDescription); err != nil {
		return nil, true, fmt.Errorf("unable to decode test.yaml: %w", err)
	}

	inputDir := filepath.Join(path, "input-dir")
	inputDirInfo, err := os.Lstat(inputDir)
	if err != nil {
		return nil, true, fmt.Errorf("unable to read inputDir: %w", err)
	}
	if !inputDirInfo.IsDir() {
		return nil, true, fmt.Errorf("input-dir must be a directory")
	}

	outputDir := filepath.Join(path, "expected-output")
	outputDirInfo, err := os.Lstat(outputDir)
	if err != nil {
		return nil, true, fmt.Errorf("unable to read inputDir: %w", err)
	}
	if !outputDirInfo.IsDir() {
		return nil, true, fmt.Errorf("input-dir must be a directory")
	}

	return &TestOptions{
		Description:   *testDescription,
		TestDirectory: path,
	}, true, nil
}
