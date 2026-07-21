package ports

// ResourceLimits holds the optional per-service Docker resource constraints.
type ResourceLimits struct {
	RAM string  // e.g. "512m", "" means no limit
	CPU float64 // e.g. 0.5, 0 means no limit
}

// HasLimits returns true if at least one limit is set.
func (rl ResourceLimits) HasLimits() bool {
	return rl.RAM != "" || rl.CPU > 0
}

// ComposeService represents a generic docker-compose service
type ComposeService struct {
	Name          string
	Image         string
	ContainerName string
	DependsOn     []string
	Environment   map[string]string
	Ports         []string
	Volumes       []string
	Networks      map[string][]string // map network name to list of aliases
	ExtraHosts    []string
	Limits        ResourceLimits
	Build         *ComposeBuild
}

// ComposeBuild holds docker-compose build instructions
type ComposeBuild struct {
	Context    string
	Dockerfile string
	Args       map[string]string
}

type ComposeBuilder interface {
	AddService(svc ComposeService)
	AddNetwork(name string, subnet string)
	Build() string
}

type ComposeBuilderFactory func() ComposeBuilder

// KubeContainer represents a generic kubernetes container within a pod
type KubeContainer struct {
	Name         string
	Image        string
	Ports        []KubePort
	Env          map[string]string
	VolumeMounts []KubeVolumeMount
	Command      []string
}

type KubePort struct {
	ContainerPort int
	Protocol      string
}

type KubeVolumeMount struct {
	Name      string
	MountPath string
	SubPath   string
}

type KubeVolume struct {
	Name      string
	ConfigMap string
	HostPath  string
	Type      string
}

type KubeServicePort struct {
	Name       string
	Port       int
	TargetPort int
	Protocol   string
}

type KubernetesBuilder interface {
	AddConfigMap(name string, files map[string]string)
	AddDeployment(name string, containers []KubeContainer, volumes []KubeVolume, nodeLabel string, hostNetwork bool)
	AddService(name string, selector map[string]string, ports []KubeServicePort, isLoadBalancer bool)
	Build() string
}

type KubernetesBuilderFactory func() KubernetesBuilder
