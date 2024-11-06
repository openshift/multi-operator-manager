package libraryinputresources

import (
	"context"
	"os"
	"path"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	"sigs.k8s.io/yaml"
)

func TestGetRequiredResourcesFromMustGather(t *testing.T) {
	// this exists to make it easy to generate output for the unit tests
	writeActualContent := false

	content, err := os.ReadDir("test-data")
	if err != nil {
		t.Fatal(err)
	}

	for _, currTestDir := range content {
		ctx := context.Background()
		testName := currTestDir.Name()
		t.Run(testName, func(t *testing.T) {
			testPertinentResources := path.Join("test-data", currTestDir.Name(), "input-resources.yaml")
			pertinentResourcesBytes, err := os.ReadFile(testPertinentResources)
			if err != nil {
				t.Fatal(err)
			}
			pertinentResources := &InputResources{}
			if err := yaml.Unmarshal(pertinentResourcesBytes, &pertinentResources); err != nil {
				t.Fatal(err)
			}

			mustGatherDirPath := path.Join("test-data", currTestDir.Name(), "input-dir")
			expectedDirPath := path.Join("test-data", currTestDir.Name(), "expected-output")

			if writeActualContent {
				err = WriteRequiredInputResourcesFromMustGather(ctx, pertinentResources, mustGatherDirPath, expectedDirPath)
				if err != nil {
					t.Fatal(err)
				}
				return
			}

			actualPertinentResources, err := GetRequiredInputResourcesFromMustGather(ctx, pertinentResources, mustGatherDirPath)
			if err != nil {
				t.Fatal(err)
			}

			expectedPertinentResources, err := LenientResourcesFromDirRecursive(expectedDirPath)
			if err != nil {
				t.Fatal(err)
			}

			differences := EquivalentResources("pruned", expectedPertinentResources, actualPertinentResources)
			if len(differences) > 0 {
				t.Log(strings.Join(differences, "\n"))
				t.Errorf("expected results mismatch %d times with actual results", len(differences))
			}
		})
	}
}

func TestUniqueResourceSet(t *testing.T) {
	name1 := "audit"
	name2 := "audit-revision-1"
	namespace1 := "openshift-authentication"
	namespace2 := "openshift-cluster-csi-drivers"
	testCases := []struct {
		name     string
		existing []*Resource
		input    []*Resource
		expected []*Resource
	}{
		{
			name:     "No input to an empty slice",
			existing: nil,
			input:    nil,
			expected: nil,
		},
		{
			name:     "No input to an existing slice",
			existing: makeResourcesWithNames(namespace1, name1),
			input:    nil,
			expected: makeResourcesWithNames(namespace1, name1),
		},
		{
			name:     "Adds a single item to an empty slice",
			existing: nil,
			input:    makeResourcesWithNames(namespace1, name1),
			expected: makeResourcesWithNames(namespace1, name1),
		},
		{
			name:     "No duplicates",
			existing: makeResourcesWithNames(namespace1, name1),
			input:    makeResourcesWithNames(namespace1, name2),
			expected: makeResourcesWithNames(namespace1, name1, name2),
		},
		{
			name:     "With duplicates",
			existing: makeResourcesWithNames(namespace1, name1),
			input:    makeResourcesWithNames(namespace1, name1, name2, name1),
			expected: makeResourcesWithNames(namespace1, name1, name2),
		},
		{
			name:     "Only duplicates",
			existing: makeResourcesWithNames(namespace1, name1),
			input:    makeResourcesWithNames(namespace1, name1, name1, name1),
			expected: makeResourcesWithNames(namespace1, name1),
		},
		{
			name:     "No duplicates if same name but different namespace",
			existing: makeResourcesWithNames(namespace1, name1),
			input:    makeResourcesWithNames(namespace2, name1),
			expected: func() []*Resource {
				return append(
					makeResourcesWithNames(namespace1, name1),
					makeResourcesWithNames(namespace2, name1)...,
				)
			}(),
		},
		{
			name: "With duplicates in a single namespace",
			existing: func() []*Resource {
				return append(
					makeResourcesWithNames(namespace1, name1),
					makeResourcesWithNames(namespace2, name1)...,
				)
			}(),
			input: makeResourcesWithNames(namespace1, name1),
			expected: func() []*Resource {
				return append(
					makeResourcesWithNames(namespace1, name1),
					makeResourcesWithNames(namespace2, name1)...,
				)
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			instances := NewUniqueResourceSet(tc.existing...)
			instances.Insert(tc.input...)
			result := instances.List()
			if len(result) != len(tc.expected) {
				t.Errorf("expected %d items, got %d", len(tc.expected), len(result))
			}
			if !equality.Semantic.DeepEqual(result, tc.expected) {
				t.Errorf(diff.ObjectDiff(tc.expected, result))
			}
		})
	}
}

func makeResourcesWithNames(namespace string, names ...string) []*Resource {
	resources := []*Resource{}
	for _, name := range names {
		content := &unstructured.Unstructured{}
		content.SetName(name)
		content.SetNamespace(namespace)
		resources = append(resources, &Resource{
			ResourceType: schema.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "configmaps",
			},
			Content: content,
		})
	}
	return resources
}
