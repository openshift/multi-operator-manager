package testapplyconfiguration

type TestDescription struct {
	BinaryName string   `json:"binaryName"`
	TestName   string   `json:"testName"`
	TestType   TestType `json:"testType"`
}

type TestType string

var (
	TestTypeApplyConfiguration TestType = "ApplyConfiguration"
)
