package librarydependson

type PertinentResources struct {
	// configurationResources is the list of resources that describe how to configure the operand.
	// on hypershift, these will always be on the management cluster only
	ConfigurationResources ResourceList `json:"configurationResources"`

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
	ConfigurationServerResources ResourceList `json:"configurationServerResources"`
	ManagementServerResources    ResourceList `json:"managementServerResources"`
	GuestServerResources         ResourceList `json:"guestServerResources"`
}

type ExactResource struct {
	ResourceTypeIdentifier `json:",inline"`

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
	ResourceTypeIdentifier `json:",inline"`

	// may have multiple matches
	// TODO CEL may be more appropriate
	ContainerJSONPath string `json:"containerJSONPath"`
	NamespaceField    string `json:"namespaceField"`
	NameField         string `json:"nameField"`
}

type ImplicitNamespacedReference struct {
	ResourceTypeIdentifier `json:",inline"`

	Namespace string `json:"namespace"`
	// may have multiple matches
	// TODO CEL may be more appropriate
	NameJSONPath string `json:"nameJSONPath"`
}

type ClusterScopedReference struct {
	ResourceTypeIdentifier `json:",inline"`

	// may have multiple matches
	// TODO CEL may be more appropriate
	NameJSONPath string `json:"nameJSONPath"`
}

type ResourceTypeIdentifier struct {
	Group string `json:"group"`
	// version is very important because it must match the version of serialization that your operator expects.
	// All Group,Resource tuples must use the same Version.
	Version  string `json:"version"`
	Resource string `json:"resource"`
}
