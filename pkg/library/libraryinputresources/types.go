package libraryinputresources

// InputResources contains the items that an operator needs to make a decision about what needs to be create,
// modified, or removed.
type InputResources struct {
	// applyConfigurationResources are the list of resources used as input to the apply-configuration command.
	// It is the responsibility of the MOM to determine where the inputs come from.
	ApplyConfigurationResources ResourceList `json:"applyConfigurationResources"`

	// operandResources is the list of resources that are important for determining check-health
	OperandResources OperandResourceList `json:"operandResources"`
}

type ResourceList struct {
	ExactResources []ExactResource `json:"exactResources"`

	// use resourceReferences when one resource (apiserver.config.openshift.io/cluster) refers to another resource
	// like a secret (.spec.servingCerts.namedCertificates[*].servingCertificates.name).
	ResourceReference []ResourceReference `json:"resourceReferences"`
}

type OperandResourceList struct {
	ConfigurationResources ResourceList `json:"configurationResources"`
	ManagementResources    ResourceList `json:"managementResources"`
	UserWorkloadResources  ResourceList `json:"userWorkloadResources"`
}

type ExactResource struct {
	DependsOnResourceTypeIdentifier `json:",inline"`

	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type ResourceReference struct {
	// TODO determine if we need the ability to select multiple containing resources.  I don’t think we’ll need to given the shape of our configuration.
	ReferringResource ExactResource `json:"referringResource"`

	Type string `json:"type"`

	ExplicitNamespacedReference *ExplicitNamespacedReference `json:"explicitNamespacedReference"`
	ImplicitNamespacedReference *ImplicitNamespacedReference `json:"implicitNamespacedReference"`
	ClusterScopedReference      *ClusterScopedReference      `json:"clusterScopedReference"`
}

type ExplicitNamespacedReference struct {
	DependsOnResourceTypeIdentifier `json:",inline"`

	// may have multiple matches
	// TODO CEL may be more appropriate
	ContainerJSONPath string `json:"containerJSONPath"`
	NamespaceField    string `json:"namespaceField"`
	NameField         string `json:"nameField"`
}

type ImplicitNamespacedReference struct {
	DependsOnResourceTypeIdentifier `json:",inline"`

	Namespace string `json:"namespace"`
	// may have multiple matches
	// TODO CEL may be more appropriate
	NameJSONPath string `json:"nameJSONPath"`
}

type ClusterScopedReference struct {
	DependsOnResourceTypeIdentifier `json:",inline"`

	// may have multiple matches
	// TODO CEL may be more appropriate
	NameJSONPath string `json:"nameJSONPath"`
}

type DependsOnResourceTypeIdentifier struct {
	Group string `json:"group"`
	// version is very important because it must match the version of serialization that your operator expects.
	// All Group,Resource tuples must use the same Version.
	Version  string `json:"version"`
	Resource string `json:"resource"`
}
