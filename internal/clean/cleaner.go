package clean

import (
	"os"

	"broom/internal/scan"
)

// Item deletes a single scan.Item from disk.
func Item(item scan.Item) error {
	if item.CleanFunc != nil {
		return item.CleanFunc()
	}
	return os.RemoveAll(item.Path)
}
