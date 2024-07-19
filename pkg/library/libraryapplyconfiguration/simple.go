package libraryapplyconfiguration

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

type ClusterApplyResult interface {
	GetClusterType() ClusterType

	ToApply() ([]Resource, error)
	ToApplyStatus() ([]Resource, error)
	ToCreate() ([]Resource, error)
	ToUpdate() ([]Resource, error)
	ToUpdateStatus() ([]Resource, error)
	ToDelete() ([]Resource, error)
}

type ApplyConfiguration struct {
	DesiredConfigurationCluster ClusterApplyResult
	DesiredManagementCluster    ClusterApplyResult
	DesiredUserWorkloadCluster  ClusterApplyResult
}

func (s *ApplyConfiguration) Validate() error {
	errs := []error{}

	if s == nil {
		return fmt.Errorf("ApplyConfiguration is required")
	}
	if s.DesiredConfigurationCluster == nil {
		errs = append(errs, fmt.Errorf("DesiredConfigurationCluster info is required even if empty"))
	} else {
		if s.DesiredConfigurationCluster.GetClusterType() != ClusterTypeConfiguration {
			errs = append(errs, fmt.Errorf("DesiredConfigurationCluster.GetClusterType must be %v", ClusterTypeConfiguration))
		}
	}
	if s.DesiredManagementCluster == nil {
		errs = append(errs, fmt.Errorf("DesiredManagementCluster info is required even if empty"))
	} else {
		if s.DesiredManagementCluster.GetClusterType() != ClusterTypeManagement {
			errs = append(errs, fmt.Errorf("DesiredManagementCluster.GetClusterType must be %v", ClusterTypeManagement))
		}
	}
	if s.DesiredUserWorkloadCluster == nil {
		errs = append(errs, fmt.Errorf("DesiredUserWorkloadCluster info is required even if empty"))
	} else {
		if s.DesiredUserWorkloadCluster.GetClusterType() != ClusterTypeUserWorkload {
			errs = append(errs, fmt.Errorf("DesiredUserWorkloadCluster.GetClusterType must be %v", ClusterTypeUserWorkload))
		}
	}

	return errors.Join(errs...)
}

type ClusterType string

var (
	ClusterTypeConfiguration ClusterType = "Configuration"
	ClusterTypeManagement    ClusterType = "Management"
	ClusterTypeUserWorkload  ClusterType = "UserWorkload"
	KnownClusterTypes                    = sets.New(ClusterTypeConfiguration, ClusterTypeManagement, ClusterTypeUserWorkload)
)

type SimpleClusterApplyResult struct {
	ClusterType ClusterType

	Apply        []Resource
	ApplyStatus  []Resource
	Create       []Resource
	Update       []Resource
	UpdateStatus []Resource
	Delete       []Resource
}

type Resource struct {
	Filename string
	Content  *unstructured.Unstructured
}

func ResourceFromFile(location string) (*Resource, error) {
	content, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("unable to read %q: %w", location, err)
	}

	ret, _, jsonErr := unstructured.UnstructuredJSONScheme.Decode(content, nil, &unstructured.Unstructured{})
	if jsonErr != nil {
		// try to see if it's yaml
		jsonString, err := yaml.YAMLToJSON(content)
		if err != nil {
			return nil, fmt.Errorf("unable to decode %q as json: %w", location, jsonErr)
		}
		ret, _, err = unstructured.UnstructuredJSONScheme.Decode(jsonString, nil, &unstructured.Unstructured{})
		if err != nil {
			return nil, fmt.Errorf("unable to decode %q as yaml: %w", location, err)
		}
	}

	return &Resource{
		Filename: filepath.Base(location),
		Content:  ret.(*unstructured.Unstructured),
	}, nil
}

func (s *SimpleClusterApplyResult) GetClusterType() ClusterType {
	return s.ClusterType
}

func (s *SimpleClusterApplyResult) ToApply() ([]Resource, error) {
	return s.Apply, nil
}

func (s *SimpleClusterApplyResult) ToApplyStatus() ([]Resource, error) {
	return s.ApplyStatus, nil
}

func (s *SimpleClusterApplyResult) ToCreate() ([]Resource, error) {
	return s.Create, nil
}

func (s *SimpleClusterApplyResult) ToUpdate() ([]Resource, error) {
	return s.Update, nil
}

func (s *SimpleClusterApplyResult) ToUpdateStatus() ([]Resource, error) {
	return s.UpdateStatus, nil
}

func (s *SimpleClusterApplyResult) ToDelete() ([]Resource, error) {
	return s.Delete, nil
}
