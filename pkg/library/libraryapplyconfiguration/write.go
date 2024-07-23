package libraryapplyconfiguration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

func WriteApplyConfiguration(desiredApplyConfiguration *ApplyConfiguration, outputDirectory string) error {
	errs := []error{}

	clusterTypeDir := filepath.Join(outputDirectory, string(desiredApplyConfiguration.DesiredConfigurationCluster.GetClusterType()))
	if err := WriteClusterApplyResult(desiredApplyConfiguration.DesiredConfigurationCluster, clusterTypeDir); err != nil {
		errs = append(errs, err)
	}

	clusterTypeDir = filepath.Join(outputDirectory, string(desiredApplyConfiguration.DesiredManagementCluster.GetClusterType()))
	if err := WriteClusterApplyResult(desiredApplyConfiguration.DesiredManagementCluster, clusterTypeDir); err != nil {
		errs = append(errs, err)
	}

	clusterTypeDir = filepath.Join(outputDirectory, string(desiredApplyConfiguration.DesiredUserWorkloadCluster.GetClusterType()))
	if err := WriteClusterApplyResult(desiredApplyConfiguration.DesiredUserWorkloadCluster, clusterTypeDir); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func WriteClusterApplyResult(desiredApplyConfiguration ClusterApplyResult, outputDirectory string) error {
	applyDir := filepath.Join(outputDirectory, "Apply")
	applyStatusDir := filepath.Join(outputDirectory, "ApplyStatus")
	createDir := filepath.Join(outputDirectory, "Create")
	updateDir := filepath.Join(outputDirectory, "Update")
	updateStatusDir := filepath.Join(outputDirectory, "UpdateStatus")
	deleteDir := filepath.Join(outputDirectory, "Delete")
	allVerbDirs := []string{applyDir, applyStatusDir, createDir, updateDir, updateStatusDir, deleteDir}

	errs := []error{}
	for _, verbDir := range allVerbDirs {
		if err := os.MkdirAll(verbDir, 0755); err != nil && !os.IsExist(err) {
			errs = append(errs, fmt.Errorf("failed creating %q: %w", verbDir, err))
			continue
		}
		verb := filepath.Base(verbDir)

		currResourceList := []*Resource{}
		var err error
		switch verb {
		case "Apply":
			currResourceList, err = desiredApplyConfiguration.ToApply()
		case "ApplyStatus":
			currResourceList, err = desiredApplyConfiguration.ToApplyStatus()
		case "Create":
			currResourceList, err = desiredApplyConfiguration.ToCreate()
		case "Update":
			currResourceList, err = desiredApplyConfiguration.ToUpdate()
		case "UpdateStatus":
			currResourceList, err = desiredApplyConfiguration.ToUpdateStatus()
		case "Delete":
			currResourceList, err = desiredApplyConfiguration.ToDelete()
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("failed getting resources to serialize %q: %w", verb, err))
			continue
		}

		for _, currResource := range currResourceList {
			filename := filepath.Join(verbDir, currResource.Filename)

			content, err := yaml.Marshal(currResource.Content)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed encoding %q: %w", filename, err))
				continue
			}
			if err := os.WriteFile(filename, content, 0644); err != nil {
				errs = append(errs, fmt.Errorf("failed writing %q: %w", filename, err))
				continue
			}
		}
		if len(currResourceList) == 0 {
			filename := filepath.Join(verbDir, ".gitkeep")
			content := []byte("this file exists to get the empty directory in git\n")
			if err := os.WriteFile(filename, content, 0644); err != nil {
				errs = append(errs, fmt.Errorf("failed writing %q: %w", filename, err))
				continue
			}
		}
	}

	return errors.Join(errs...)
}
