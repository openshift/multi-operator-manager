package libraryapplyconfiguration

import (
	"embed"
	"github.com/google/go-cmp/cmp"
	"io/fs"
	"strings"
	"testing"
)

//go:embed testdata
var packageTestData embed.FS

func must[A any](in A, err error) A {
	if err != nil {
		panic(err)
	}
	return in
}

func TestEquivalentApplyConfigurationResultIgnoringEvents(t *testing.T) {
	type args struct {
		lhsFS fs.FS
		rhsFS fs.FS
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "honor fieldmanager in diff",
			args: args{
				lhsFS: must(fs.Sub(packageTestData, "testdata/need-useful-diff/lhs")),
				rhsFS: must(fs.Sub(packageTestData, "testdata/need-useful-diff/rhs")),
			},
			want: []string{
				`Management: rhs is missing equivalent request for fieldManager=openshift-authentication-Metadata controllerInstanceName=TODO-metadataController: ApplyStatus-Authentication.v1.operator.openshift.io/cluster[]`,
				`Management: mutation: ApplyStatus-Authentication.v1.operator.openshift.io/cluster[], fieldManager=openshift-authentication-PayloadConfig, controllerInstanceName=TODO-payloadConfigController, rhs[1]: body diff:   []uint8(
          	"""
          	... // 7 identical lines
          	  conditions:
          	  - lastTransitionTime: "2024-10-14T22:38:20Z"
        - 	    message: 'Unable to get cluster authentication config: authentications.operator.openshift.io
        - 	      "cluster" not found'
        - 	    reason: GetFailed
        - 	    status: "True"
        + 	    message: ""
        + 	    reason: ""
        + 	    status: "False"
          	    type: OAuthConfigDegraded
          	  - lastTransitionTime: "2024-10-14T22:38:20Z"
          	... // 21 identical lines
          	"""
          )
`,
				`Management: lhs is missing equivalent request for fieldManager=oauth-server-ResourceSync controllerInstanceName=TODO-resourceSyncer: ApplyStatus-Authentication.v1.operator.openshift.io/cluster[]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lhs, err := NewApplyConfigurationResultFromDirectory(tt.args.lhsFS, "test", nil)
			if err != nil {
				t.Fatal(err)
			}
			rhs, err := NewApplyConfigurationResultFromDirectory(tt.args.rhsFS, "test", nil)
			if err != nil {
				t.Fatal(err)
			}

			got := EquivalentApplyConfigurationResultIgnoringEvents(lhs, rhs)
			if len(got) != len(tt.want) {
				t.Fatal(got)
			}
			for i := range got {
				// workaround the intentional whitespace trick in cmp.Diff because we want stable results, but don't care if we have check format every once in a while: https://github.com/google/go-cmp/issues/235
				currActuals := strings.Split(strings.ReplaceAll(got[i], "\u00a0", " "), "\n")
				currExpecteds := strings.Split(strings.ReplaceAll(tt.want[i], "\u00a0", " "), "\n")
				if len(currActuals) != len(currExpecteds) {
					t.Fatal(got[i])
				}

				for j := range currActuals {
					currActual := strings.TrimSpace(currActuals[j])
					currExpected := strings.TrimSpace(currExpecteds[j])
					if currActual != currExpected {
						t.Error(cmp.Diff(currActual, currExpected))
					}
				}

			}
		})
	}
}
