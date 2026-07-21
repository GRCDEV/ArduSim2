package ports

import (
	"context"
)

// Server defines the contract for any server
type Server interface {
	Start(ctx context.Context) error
	Stop() error
}
