package testapplyconfiguration

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type TestDescription struct {
	BinaryName   string    `json:"binaryName"`
	TestName     string    `json:"testName"`
	Description  string    `json:"description"`
	TestType     TestType  `json:"testType"`
	DesiredError ErrorType `json:"desiredError,omitempty"`
	// Now is the time to use when invoking the apply-configuration command.  This is commonly used so that output
	// for conditions is stable
	Now metav1.Time `json:"now"`
	// ControllersToRun holds an optional list of controller names to run.
	// By default, all controllers are run.
	ControllersToRun []string `json:"controllersToRun,omitempty"`
}

type TestType string

var (
	TestTypeApplyConfiguration TestType = "ApplyConfiguration"
	AllTestTypes                        = sets.New(TestTypeApplyConfiguration)
)

type ErrorType string

var (
	NoError       ErrorType = ""
	NonZeroReturn ErrorType = "NonZeroReturn"
)
