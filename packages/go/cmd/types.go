package cmd

// Container represents the top-level container object.
type Container struct {
	Networks    []interface{} `json:"networks"`
	Status      string        `json:"status"`
	Config      Configuration `json:"configuration"`
	StartedDate float64       `json:"startedDate"`
}

// Configuration holds the full container configuration.
type Configuration struct {
	Networks         []NetworkEntry    `json:"networks"`
	UseInit          bool              `json:"useInit"`
	Resources        Resources         `json:"resources"`
	RuntimeHandler   string            `json:"runtimeHandler"`
	SSH              bool              `json:"ssh"`
	InitProcess      InitProcess       `json:"initProcess"`
	Rosetta          bool              `json:"rosetta"`
	Sysctls          map[string]string `json:"sysctls"`
	PublishedSockets []interface{}     `json:"publishedSockets"`
	Mounts           []Mount           `json:"mounts"`
	ReadOnly         bool              `json:"readOnly"`
	Virtualization   bool              `json:"virtualization"`
	PublishedPorts   []interface{}     `json:"publishedPorts"`
	Image            Image             `json:"image"`
	Labels           map[string]string `json:"labels"`
	Platform         Platform          `json:"platform"`
	ID               string            `json:"id"`
	DNS              DNS               `json:"dns"`
}

// NetworkEntry represents a single network configuration entry.
type NetworkEntry struct {
	Options NetworkOptions `json:"options"`
	Network string         `json:"network"`
}

// NetworkOptions holds options for a network entry.
type NetworkOptions struct {
	Hostname string `json:"hostname"`
}

// Resources describes the resource limits for the container.
type Resources struct {
	CPUs          int   `json:"cpus"`
	MemoryInBytes int64 `json:"memoryInBytes"`
}

// InitProcess describes the initial process to run in the container.
type InitProcess struct {
	WorkingDirectory   string   `json:"workingDirectory"`
	Terminal           bool     `json:"terminal"`
	User               User     `json:"user"`
	SupplementalGroups []string `json:"supplementalGroups"`
	Executable         string   `json:"executable"`
	Environment        []string `json:"environment"`
	Arguments          []string `json:"arguments"`
	Rlimits            []string `json:"rlimits"`
}

// User represents the user configuration for the init process.
type User struct {
	Raw RawUser `json:"raw"`
}

// RawUser holds the raw user string identifier.
type RawUser struct {
	UserString string `json:"userString"`
}

// Mount describes a filesystem mount inside the container.
type Mount struct {
	Type        MountType `json:"type"`
	Destination string    `json:"destination"`
	Options     []string  `json:"options"`
	Source      string    `json:"source"`
}

// MountType represents the type of mount (e.g. virtiofs).
type MountType struct {
	VirtioFS *VirtioFS `json:"virtiofs,omitempty"`
}

// VirtioFS is a marker type for virtiofs mounts.
type VirtioFS struct{}

// Image holds the container image reference and descriptor.
type Image struct {
	Reference  string          `json:"reference"`
	Descriptor ImageDescriptor `json:"descriptor"`
}

// ImageDescriptor contains the OCI descriptor for the image.
type ImageDescriptor struct {
	Digest      string            `json:"digest"`
	Annotations map[string]string `json:"annotations"`
	MediaType   string            `json:"mediaType"`
	Size        int64             `json:"size"`
}

// Platform specifies the OS and architecture for the container.
type Platform struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
}

// DNS holds DNS configuration for the container.
type DNS struct {
	Nameservers   []string `json:"nameservers"`
	SearchDomains []string `json:"searchDomains"`
	Options       []string `json:"options"`
}
