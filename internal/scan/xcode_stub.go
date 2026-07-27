//go:build !darwin

package scan

import (
	"context"

	"broom/internal/platform"
)

// XcodeScanner is a no-op on non-macOS platforms.
type XcodeScanner struct{}

func (s *XcodeScanner) Name() string { return "xcode" }

func (s *XcodeScanner) Scan(_ context.Context, _ []platform.Root, _ ScanContext, _ chan<- Item) {
}
