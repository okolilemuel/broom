package platform

// Root is a directory to scan with a max recursion depth.
type Root struct {
	Path     string
	MaxDepth int
}

// CacheEntry is a named global cache directory.
type CacheEntry struct {
	Name     string
	Path     string
	Category string
}

// Platform abstracts OS-specific paths and behaviours.
type Platform interface {
	Name() string
	// SearchRoots returns directories to walk looking for project artifacts.
	SearchRoots() []Root
	// GlobalCaches returns named cache dirs to check as a whole.
	GlobalCaches() []CacheEntry
	// DiskFreeBytes returns available bytes for the given path.
	DiskFreeBytes(path string) (int64, error)
}
