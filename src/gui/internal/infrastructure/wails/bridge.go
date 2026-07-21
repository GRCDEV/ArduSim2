package wails

import (
	"context"
	"os"
	"ui/internal/ports"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type WailsBridge struct {
	ctx context.Context
}

func NewWailsBridge() *WailsBridge {
	return &WailsBridge{}
}

func (b *WailsBridge) SetContext(ctx context.Context) {
	b.ctx = ctx
}

func (b *WailsBridge) EmitEvent(name string, data ...interface{}) {
	if b.ctx != nil {
		runtime.EventsEmit(b.ctx, name, data...)
	}
}

func (b *WailsBridge) OpenDirectoryDialog(ctx context.Context, title, defaultDir string) (string, error) {
	if info, err := os.Stat(defaultDir); err != nil || !info.IsDir() {
		defaultDir = ""
	}
	
	return runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		DefaultDirectory: defaultDir,
		Title:            title,
	})
}

func (b *WailsBridge) OpenFileDialog(ctx context.Context, title string, filters []ports.FileFilter) (string, error) {
	wailsFilters := make([]runtime.FileFilter, len(filters))
	for i, f := range filters {
		wailsFilters[i] = runtime.FileFilter{
			DisplayName: f.DisplayName,
			Pattern:     f.Pattern,
		}
	}
	return runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: wailsFilters,
	})
}

func (b *WailsBridge) SaveFileDialog(ctx context.Context, title string, defaultFilename string, filters []ports.FileFilter) (string, error) {
	wailsFilters := make([]runtime.FileFilter, len(filters))
	for i, f := range filters {
		wailsFilters[i] = runtime.FileFilter{
			DisplayName: f.DisplayName,
			Pattern:     f.Pattern,
		}
	}
	return runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultFilename,
		Filters:         wailsFilters,
	})
}
