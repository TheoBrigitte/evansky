package cache

const (
	DoNotExists Status = "DoNotExists"
	UpToDate    Status = "UpToDate"
	Changed     Status = "Changed"
)

type Status string

// Cache represents a cache entry.
// Cache entries are store in user cache directory (see os.UserCacheDir) under ProjectCacheDir.
// Each cache entry is a folder where the name is the hash of the path.
// e.g. md5(test) = 098f6bcd4621d373cade4e832627b4f6
//      ./test -> /path/to/cache/098f6bcd4621d373cade4e832627b4f6
type Cache struct {
	dir           string
	fileschecksum string
	scan          string
}
