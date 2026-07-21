package domain

// VolumeMount represents a directory or file mapping from the simulation's
// host 'resources' directory to a target path inside a container.
type VolumeMount struct {
	HostPath      string
	ContainerPath string
}
