package libraryproduces

type ProducedResources struct {
	ConfigurationServerResources ResourceList `json:"configurationServerResources"`
	ManagementServerResources    ResourceList `json:"managementServerResources"`
	GuestServerResources         ResourceList `json:"guestServerResources"`
}

type ResourceList struct {
	ExactResources []ExactResource `json:"exactResources"`

	GeneratedNameResources []GeneratedNameResource `json:"generatedNameResource"`

	// TODO I bet this covers 95% of what we need, but maybe we need label selector.
	// I'm a solid -1 on "pattern" based selection. We select in kube based on label selectors.
}

type ExactResource struct {
	ResourceTypeIdentifier `json:",inline"`

	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
}

type GeneratedNameResource struct {
	ResourceTypeIdentifier `json:",inline"`

	Namespace     string `json:"namespace,omitempty"`
	GeneratedName string `json:"name"`
}

type ResourceTypeIdentifier struct {
	Group string `json:"group"`
	// version is very important because it must match the version of serialization that your operator expects.
	// All Group,Resource tuples must use the same Version.
	Version  string `json:"version"`
	Resource string `json:"resource"`
}
