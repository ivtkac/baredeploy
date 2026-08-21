// Package version holds build metadata injected at link time via
// -ldflags (see Makefile).
package version

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns a human-readable version summary.
func String() string {
	return Version + " (commit " + Commit + ", built " + Date + ")"
}
