package libraryapplyconfiguration

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
)

func TestValidateControllers(t *testing.T) {
	scenarios := []struct {
		name                      string
		knownControllers          []string
		controllersToRunFromFlags []string
		expectedErrors            []string
	}{
		{
			name:                      "all known controllers",
			knownControllers:          []string{"controllerA", "controllerB", "controllerC"},
			controllersToRunFromFlags: []string{"controllerA", "controllerB"},
			expectedErrors:            nil,
		},
		{
			name:                      "unknown controller",
			knownControllers:          []string{"controllerA", "controllerB", "controllerC"},
			controllersToRunFromFlags: []string{"controllerD"},
			expectedErrors:            []string{`"controllerD" is not in the list of known controllers`},
		},
		{
			name:                      "mix of known and unknown controllers",
			knownControllers:          []string{"controllerA", "controllerB", "controllerC"},
			controllersToRunFromFlags: []string{"controllerA", "controllerD"},
			expectedErrors:            []string{`"controllerD" is not in the list of known controllers`},
		},
		{
			name:                      "wildcard",
			knownControllers:          []string{"controllerA", "controllerB", "controllerC"},
			controllersToRunFromFlags: []string{"*"},
			expectedErrors:            nil,
		},
		{
			name:                      "prefixed known controller",
			knownControllers:          []string{"controllerA", "controllerB", "controllerC"},
			controllersToRunFromFlags: []string{"-controllerA"},
			expectedErrors:            nil,
		},
		{
			name:                      "prefixed unknown controller",
			knownControllers:          []string{"controllerA", "controllerB", "controllerC"},
			controllersToRunFromFlags: []string{"-controllerD"},
			expectedErrors:            []string{`"controllerD" is not in the list of known controllers`},
		},
		{
			name:                      "mix of prefixed and known controllers",
			knownControllers:          []string{"controllerA", "controllerB", "controllerC"},
			controllersToRunFromFlags: []string{"controllerA", "-controllerD"},
			expectedErrors:            []string{`"controllerD" is not in the list of known controllers`},
		},
		{
			name:                      "empty list of controllers",
			knownControllers:          []string{"controllerA", "controllerB", "controllerC"},
			controllersToRunFromFlags: []string{},
			expectedErrors:            nil,
		},
		{
			name:                      "no known controllers",
			knownControllers:          []string{},
			controllersToRunFromFlags: []string{"controllerA"},
			expectedErrors:            []string{`"controllerA" is not in the list of known controllers`},
		},
	}

	for _, tt := range scenarios {
		t.Run(tt.name, func(t *testing.T) {
			allKnownControllersSet := sets.NewString(tt.knownControllers...)

			errs := validateControllersFromFlags(allKnownControllersSet, tt.controllersToRunFromFlags)

			var actualErrors []string
			for _, err := range errs {
				actualErrors = append(actualErrors, err.Error())
			}

			if len(actualErrors) != len(tt.expectedErrors) {
				t.Errorf("expected %d errors, got %d errors", len(tt.expectedErrors), len(actualErrors))
			}

			for i, expectedErr := range tt.expectedErrors {
				if actualErrors[i] != expectedErr {
					t.Errorf("expected error %q, got %q", expectedErr, actualErrors[i])
				}
			}
		})
	}
}

func TestIsControllerEnabled(t *testing.T) {
	scenarios := []struct {
		name                 string
		controllerToRun      string
		controllersFromFlags []string
		expected             bool
	}{
		{
			name:                 "explicitly enabled controller",
			controllerToRun:      "controllerA",
			controllersFromFlags: []string{"controllerA", "controllerB"},
			expected:             true,
		},
		{
			name:                 "explicitly disabled controller",
			controllerToRun:      "controllerA",
			controllersFromFlags: []string{"-controllerA", "controllerB"},
			expected:             false,
		},
		{
			name:                 "wildcard enabled",
			controllerToRun:      "controllerA",
			controllersFromFlags: []string{"*"},
			expected:             true,
		},
		{
			name:                 "wildcard with explicitly disabled controller",
			controllerToRun:      "controllerA",
			controllersFromFlags: []string{"*", "-controllerA"},
			expected:             false,
		},
		{
			name:                 "controller not in list",
			controllerToRun:      "controllerC",
			controllersFromFlags: []string{"controllerA", "controllerB"},
			expected:             false,
		},
		{
			name:                 "wildcard with no controller specified",
			controllerToRun:      "controllerA",
			controllersFromFlags: []string{"*"},
			expected:             true,
		},
		{
			name:                 "controller not in list with no wildcard",
			controllerToRun:      "controllerC",
			controllersFromFlags: []string{"controllerA", "controllerB"},
			expected:             false,
		},
		{
			name:                 "explicitly disabled controller with no wildcard",
			controllerToRun:      "controllerA",
			controllersFromFlags: []string{"-controllerA"},
			expected:             false,
		},
		{
			name:                 "empty controller list",
			controllerToRun:      "controllerA",
			controllersFromFlags: []string{},
			expected:             false,
		},
	}

	for _, tt := range scenarios {
		t.Run(tt.name, func(t *testing.T) {
			result := isControllerEnabled(tt.controllerToRun, tt.controllersFromFlags)

			if result != tt.expected {
				t.Errorf("for controller: %q, expected the target method to return: %v but got %v", tt.controllerToRun, tt.expected, result)
			}
		})
	}
}
