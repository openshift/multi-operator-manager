package libraryinputresources

import (
	"reflect"
	"testing"
)

func Test_validateInputResources(t *testing.T) {
	type args struct {
		obj *InputResources
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "missing version",
			args: args{
				obj: &InputResources{
					ApplyConfigurationResources: ResourceList{
						ExactResources: []ExactResourceID{
							ExactResource("", "", "secrets", "foo", "bar"),
						},
					},
					OperandResources: OperandResourceList{},
				},
			},
			want: []string{
				"applyConfigurationResources.exactResources[0].version: Required value: must be present",
			},
		},
		{
			name: "missing namespace",
			args: args{
				obj: &InputResources{
					ApplyConfigurationResources: ResourceList{
						ExactResources: []ExactResourceID{
							ExactResource("", "v1", "secrets", "", "bar"),
						},
					},
					OperandResources: OperandResourceList{},
				},
			},
			want: []string{},
		},
		{
			name: "bad jsonpath",
			args: args{
				obj: &InputResources{
					ApplyConfigurationResources: ResourceList{},
					OperandResources: OperandResourceList{
						ConfigurationResources: ResourceList{
							ResourceReference: []ResourceReference{
								{
									ReferringResource: ExactResource("", "", "secrets", "foo", "bar"),
									Type:              ImplicitNamespacedReferenceType,
									ImplicitNamespacedReference: &ImplicitNamespacedReference{
										InputResourceTypeIdentifier: SecretIdentifierType(),
										Namespace:                   "openshift-config",
										NameJSONPath:                "please DON'T compile[AND foo]",
									},
								},
							},
						},
					},
				},
			},
			want: []string{
				`operandResources.configurationResources.resourceReferences[0].referringResource.version: Required value: must be present`,
				`operandResources.configurationResources.resourceReferences[0].implicitNamespacedReference.nameJSONPath: Invalid value: "please DON'T compile[AND foo]": parsing error: please DON'T compile[AND foo]	:1:8 - 1:11 unexpected Ident while scanning operator`,
			},
		},
		{
			name: "good jsonpath",
			args: args{
				obj: &InputResources{
					ApplyConfigurationResources: ResourceList{},
					OperandResources: OperandResourceList{
						ConfigurationResources: ResourceList{
							ResourceReference: []ResourceReference{
								{
									ReferringResource: ExactSecret("foo", "bar"),
									Type:              ImplicitNamespacedReferenceType,
									ImplicitNamespacedReference: &ImplicitNamespacedReference{
										InputResourceTypeIdentifier: SecretIdentifierType(),
										Namespace:                   "openshift-config",
										NameJSONPath:                `$.spec.componentRoutes[?(@.name == "my-route" && @.namespace == "openshift-authentication")].servingCertKeyPairSecret.name`,
									},
								},
							},
						},
					},
				},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateInputResources(tt.args.obj)
			actualStrings := []string{}
			for _, curr := range errs {
				actualStrings = append(actualStrings, curr.Error())
			}
			if !reflect.DeepEqual(actualStrings, tt.want) {
				t.Errorf("validateInputResources() = %v", actualStrings)
			}
		})
	}
}
