package domain

// ServiceType describes an available algorithm/service discovered on disk.
type ServiceType struct {
	ID         string `json:"id"`
	FolderName string `json:"folderName"`
	Title      string `json:"title"`
	SchemaRaw  string `json:"schemaRaw"`
}
