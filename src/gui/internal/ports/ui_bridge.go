package ports

import "context"

// UIBridge abstracts UI-specific operations like event emission and dialogs.
type UIBridge interface {
	SetContext(ctx context.Context)
	EmitEvent(name string, data ...interface{})
	OpenDirectoryDialog(ctx context.Context, title, defaultDir string) (string, error)
	OpenFileDialog(ctx context.Context, title string, filters []FileFilter) (string, error)
	SaveFileDialog(ctx context.Context, title string, defaultFilename string, filters []FileFilter) (string, error)
}

// FileFilter defines a file extension filter for dialogs.
type FileFilter struct {
	DisplayName string
	Pattern     string
}
