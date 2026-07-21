package domain

// DeployedService represents a user-selected service assigned to a UAV,
// including its resolved runtime configuration.
type DeployedService struct {
	InstanceId   string                 `json:"instanceId"`
	ServiceId    string                 `json:"serviceId"`
	FolderName   string                 `json:"folderName"`
	ServiceTitle string                 `json:"serviceTitle"`
	Config       map[string]interface{} `json:"config"`
	NodeLabel    string                 `json:"nodeLabel,omitempty"`
}
