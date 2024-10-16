package testapplyconfiguration

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type TestDescription struct {
	BinaryName string   `json:"binaryName"`
	TestName   string   `json:"testName"`
	TestType   TestType `json:"testType"`
	// Now is the time to use when invoking the apply-configuration command.  This is commonly used so that output
	// for conditions is stable
	Now metav1.Time `json:"now"`
}

type TestType string

var (
	TestTypeApplyConfiguration TestType = "ApplyConfiguration"
)
