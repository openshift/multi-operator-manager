package libraryapplyconfiguration

import (
	"errors"
	"fmt"
	"io/fs"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"path"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

type ClusterApplyResult interface {
	GetClusterType() ClusterType

	ToApply() ([]*Resource, error)
	ToApplyStatus() ([]*Resource, error)
	ToCreate() ([]*Resource, error)
	ToUpdate() ([]*Resource, error)
	ToUpdateStatus() ([]*Resource, error)
	ToDelete() ([]*Resource, error)
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

	Apply        []*Resource
	ApplyStatus  []*Resource
	Create       []*Resource
	Update       []*Resource
	UpdateStatus []*Resource
	Delete       []*Resource
}

type Resource struct {
	Filename     string
	ResourceType schema.GroupVersionResource
	Content      *unstructured.Unstructured
}

func ResourcesFromDir(location string) ([]*Resource, error) {
	resources, err := os.ReadDir(location)
	if err != nil {
		return nil, fmt.Errorf("unable to read requested dir %q: %w", location, err)
	}

	currResourceList := []*Resource{}
	errs := []error{}
	for _, currFile := range resources {
		currLocation := filepath.Join(location, currFile.Name())
		if currFile.IsDir() {
			errs = append(errs, fmt.Errorf("unexpected directory %q, only json and yaml content is allowed", currLocation))
			continue
		}
		if currFile.Name() == ".gitkeep" {
			continue
		}
		if !strings.HasSuffix(currFile.Name(), ".yaml") && !strings.HasSuffix(currFile.Name(), ".json") {
			errs = append(errs, fmt.Errorf("unexpected file %q, only json and yaml content is allowed", currLocation))
		}
		currResource, err := ResourceFromFile(currLocation, location)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		currResourceList = append(currResourceList, currResource)
	}

	return currResourceList, errors.Join(errs...)
}

func LenientResourcesFromDirRecursive(location string) ([]*Resource, error) {
	currResourceList := []*Resource{}
	errs := []error{}
	err := filepath.WalkDir(location, func(currLocation string, currFile fs.DirEntry, err error) error {
		if err != nil {
			errs = append(errs, err)
		}

		if currFile.IsDir() {
			return nil
		}
		if !strings.HasSuffix(currFile.Name(), ".yaml") && !strings.HasSuffix(currFile.Name(), ".json") {
			return nil
		}
		currResource, err := ResourceFromFile(currLocation, location)
		if err != nil {
			return fmt.Errorf("error deserializing %q: %w", currLocation, err)
		}
		currResourceList = append(currResourceList, currResource)

		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}

	return currResourceList, errors.Join(errs...)
}

func ResourceFromFile(location, fileTrimPrefix string) (*Resource, error) {
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

	retFilename := strings.TrimPrefix(location, fileTrimPrefix)
	retFilename = strings.TrimPrefix(retFilename, "/")

	return &Resource{
		Filename: retFilename,
		Content:  ret.(*unstructured.Unstructured),
	}, nil
}

func IdentifyResource(in *Resource) string {
	gvkString := fmt.Sprintf("%s.%s.%s/%s[%s]", in.Content.GroupVersionKind().Kind, in.Content.GroupVersionKind().Version, in.Content.GroupVersionKind().Group, in.Content.GetName(), in.Content.GetNamespace())

	return fmt.Sprintf("%s(%s)", gvkString, in.Filename)
}

func WriteResource(in *Resource, parentDir string) error {
	if len(in.Filename) == 0 {
		return fmt.Errorf("%s is missing filename", IdentifyResource(in))
	}

	dir := path.Join(parentDir, path.Dir(in.Filename))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating dir for %v: %w", IdentifyResource(in), err)
	}

	file := path.Join(parentDir, in.Filename)
	resourceYaml, err := yaml.Marshal(in.Content)
	if err != nil {
		return fmt.Errorf("error serializing %v: %w", IdentifyResource(in), err)
	}
	if err := os.WriteFile(file, resourceYaml, 0644); err != nil {
		return fmt.Errorf("error writing %v: %w", IdentifyResource(in), err)
	}

	return nil
}

func (s *SimpleClusterApplyResult) GetClusterType() ClusterType {
	return s.ClusterType
}

func (s *SimpleClusterApplyResult) ToApply() ([]*Resource, error) {
	return s.Apply, nil
}

func (s *SimpleClusterApplyResult) ToApplyStatus() ([]*Resource, error) {
	return s.ApplyStatus, nil
}

func (s *SimpleClusterApplyResult) ToCreate() ([]*Resource, error) {
	return s.Create, nil
}

func (s *SimpleClusterApplyResult) ToUpdate() ([]*Resource, error) {
	return s.Update, nil
}

func (s *SimpleClusterApplyResult) ToUpdateStatus() ([]*Resource, error) {
	return s.UpdateStatus, nil
}

func (s *SimpleClusterApplyResult) ToDelete() ([]*Resource, error) {
	return s.Delete, nil
}
