package applyconfiguration

import (
	"errors"
	"fmt"
	"github.com/deads2k/multi-operator-manager/pkg/library/libraryapplyconfiguration"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"path/filepath"
)

type ApplyConfigurationResult interface {
	Error() error
	OutputDirectory() (string, error)
	Stdout() string
	Stderr() string

	DesiredConfigurationCluster() (libraryapplyconfiguration.ClusterApplyResult, error)
	DesiredManagementCluster() (libraryapplyconfiguration.ClusterApplyResult, error)
	DesiredUserWorkloadCluster() (libraryapplyconfiguration.ClusterApplyResult, error)
}

type simpleApplyConfigurationResult struct {
	err             error
	outputDirectory string
	stdout          string
	stderr          string

	desiredConfigurationCluster libraryapplyconfiguration.ClusterApplyResult
	desiredManagementCluster    libraryapplyconfiguration.ClusterApplyResult
	desiredUserWorkloadCluster  libraryapplyconfiguration.ClusterApplyResult
}

func NewApplyConfigurationResult(outputDirectory string, execError error) (ApplyConfigurationResult, error) {
	errs := []error{}
	var err error

	stdoutContent := []byte{}
	stdoutLocation := filepath.Join(outputDirectory, "stdout.log")
	stdoutContent, err = os.ReadFile(stdoutLocation)
	if err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("failed reading %q: %w", stdoutLocation, err))
	}
	// TODO stream through and preserve first and last to avoid memory explosion
	if len(stdoutContent) > 512*1024 {
		indexToStart := len(stdoutContent) - (512 * 1024)
		stdoutContent = stdoutContent[indexToStart:]
	}

	stderrContent := []byte{}
	stderrLocation := filepath.Join(outputDirectory, "stderr.log")
	stderrContent, err = os.ReadFile(stderrLocation)
	if err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("failed reading %q: %w", stderrLocation, err))
	}
	// TODO stream through and preserve first and last to avoid memory explosion
	if len(stderrContent) > 512*1024 {
		indexToStart := len(stderrContent) - (512 * 1024)
		stderrContent = stderrContent[indexToStart:]
	}

	if execError != nil {
		return &simpleApplyConfigurationResult{
			stdout:                      string(stdoutContent),
			stderr:                      string(stderrContent),
			err:                         execError,
			outputDirectory:             outputDirectory,
			desiredConfigurationCluster: nil,
			desiredManagementCluster:    nil,
			desiredUserWorkloadCluster:  nil,
		}, execError
	}

	outputContent, err := os.ReadDir(outputDirectory)
	if err != nil {
		return nil, fmt.Errorf("unable to read output-dir content %q: %w", outputDirectory, err)
	}

	ret := &simpleApplyConfigurationResult{
		stdout:          string(stdoutContent),
		stderr:          string(stderrContent),
		outputDirectory: outputDirectory,
	}
	ret.desiredConfigurationCluster, err = NewClusterApplyResult(libraryapplyconfiguration.ClusterTypeConfiguration, outputDirectory)
	if err != nil {
		errs = append(errs, fmt.Errorf("failure building %q result: %w", libraryapplyconfiguration.ClusterTypeConfiguration, err))
	}
	ret.desiredManagementCluster, err = NewClusterApplyResult(libraryapplyconfiguration.ClusterTypeManagement, outputDirectory)
	if err != nil {
		errs = append(errs, fmt.Errorf("failure building %q result: %w", libraryapplyconfiguration.ClusterTypeManagement, err))
	}
	ret.desiredUserWorkloadCluster, err = NewClusterApplyResult(libraryapplyconfiguration.ClusterTypeUserWorkload, outputDirectory)
	if err != nil {
		errs = append(errs, fmt.Errorf("failure building %q result: %w", libraryapplyconfiguration.ClusterTypeUserWorkload, err))
	}

	// check to be sure we don't have any extra content
	for _, currContent := range outputContent {
		if currContent.Name() == "stdout.log" {
			continue
		}
		if currContent.Name() == "stderr.log" {
			continue
		}

		if !currContent.IsDir() {
			errs = append(errs, fmt.Errorf("unexpected file %q, only target cluster directories are: %v", filepath.Join(outputDirectory, currContent.Name()), sets.List(libraryapplyconfiguration.KnownClusterTypes)))
			continue
		}
		if !libraryapplyconfiguration.KnownClusterTypes.Has(libraryapplyconfiguration.ClusterType(currContent.Name())) {
			errs = append(errs, fmt.Errorf("unexpected file %q, only target cluster directories are: %v", filepath.Join(outputDirectory, currContent.Name()), sets.List(libraryapplyconfiguration.KnownClusterTypes)))
			continue
		}
	}

	ret.err = errors.Join(errs...)
	if ret.err != nil {
		// TODO may decide to disallow returning any info later
		return ret, ret.err
	}
	return ret, nil
}

func (s *simpleApplyConfigurationResult) Stdout() string {
	return s.stdout
}

func (s *simpleApplyConfigurationResult) Stderr() string {
	return s.stderr
}

func (s *simpleApplyConfigurationResult) Error() error {
	return s.err
}

func (s *simpleApplyConfigurationResult) OutputDirectory() (string, error) {
	return s.outputDirectory, nil
}

func (s *simpleApplyConfigurationResult) DesiredConfigurationCluster() (libraryapplyconfiguration.ClusterApplyResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.desiredConfigurationCluster, nil
}

func (s *simpleApplyConfigurationResult) DesiredManagementCluster() (libraryapplyconfiguration.ClusterApplyResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.desiredManagementCluster, nil
}

func (s *simpleApplyConfigurationResult) DesiredUserWorkloadCluster() (libraryapplyconfiguration.ClusterApplyResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.desiredUserWorkloadCluster, nil
}

// NewClusterApplyResult takes a standard output directory, selects the subdirectory for the clusterType, and consumes the
// content inside that directory.
// All files can be either json or yaml.
func NewClusterApplyResult(clusterType libraryapplyconfiguration.ClusterType, outputDirectory string) (libraryapplyconfiguration.ClusterApplyResult, error) {
	ret := &libraryapplyconfiguration.SimpleClusterApplyResult{
		ClusterType: clusterType,
	}

	clusterTypeDir := filepath.Join(outputDirectory, string(clusterType))
	applyDir := filepath.Join(clusterTypeDir, "Apply")
	applyStatusDir := filepath.Join(clusterTypeDir, "ApplyStatus")
	createDir := filepath.Join(clusterTypeDir, "Create")
	updateDir := filepath.Join(clusterTypeDir, "Update")
	updateStatusDir := filepath.Join(clusterTypeDir, "UpdateStatus")
	deleteDir := filepath.Join(clusterTypeDir, "Delete")
	allVerbDirs := []string{applyDir, applyStatusDir, createDir, updateDir, updateStatusDir, deleteDir}

	clusterTypeContent, err := os.ReadDir(clusterTypeDir)
	if err != nil {
		return nil, fmt.Errorf("unable to read clusterType content clusterType=%q in %q: %w", clusterType, clusterTypeDir, err)
	}

	errs := []error{}
	allowedClusterTypeSubDirectories := sets.Set[string]{}
	for _, verbDir := range allVerbDirs {
		verb := filepath.Base(verbDir)
		allowedClusterTypeSubDirectories.Insert(verb)

		currResourceList, err := libraryapplyconfiguration.ResourcesFromDir(verbDir)
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to read verb content clusterType=%q verb=%q in %q: %w", clusterType, verb, verbDir, err))
		}
		for i, currResource := range currResourceList {
			currLocation := filepath.Join(verbDir, currResource.Filename)
			currGVK := currResource.Content.GetObjectKind().GroupVersionKind()
			currNamespace := currResource.Content.GetNamespace()
			currName := currResource.Content.GetName()
			for j, otherResource := range currResourceList {
				if i == j {
					continue
				}
				otherLocation := filepath.Join(verbDir, otherResource.Filename)
				otherGVK := otherResource.Content.GetObjectKind().GroupVersionKind()
				otherNamespace := otherResource.Content.GetNamespace()
				otherName := otherResource.Content.GetName()
				if currGVK == otherGVK && currNamespace == otherNamespace && currName == otherName {
					errs = append(errs, fmt.Errorf("duplicate resource specification GVK=%v namespace=%q name=%q in %q and %q", currGVK, currNamespace, currName, currLocation, otherLocation))
				}
			}
		}

		switch verb {
		case "Apply":
			ret.Apply = currResourceList
		case "ApplyStatus":
			ret.ApplyStatus = currResourceList
		case "Create":
			ret.Create = currResourceList
		case "Update":
			ret.Update = currResourceList
		case "UpdateStatus":
			ret.UpdateStatus = currResourceList
		case "Delete":
			ret.Delete = currResourceList
		}
	}
	for _, clusterTypeSubDir := range clusterTypeContent {
		if !clusterTypeSubDir.IsDir() {
			errs = append(errs, fmt.Errorf("unexpected file %q, only verb directory content is allowed", filepath.Join(clusterTypeDir, clusterTypeSubDir.Name())))
			continue
		}
		if !allowedClusterTypeSubDirectories.Has(clusterTypeSubDir.Name()) {
			errs = append(errs, fmt.Errorf("unexpected directory %q, only verb subdirectories are allowed: %v", filepath.Join(clusterTypeDir, clusterTypeSubDir.Name()), sets.List(allowedClusterTypeSubDirectories)))
			continue
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return ret, nil
}
