// Package version holds build metadata injected via -ldflags.
package version

import "fmt"

// Set at link time via -ldflags.
var (
	Version   = "0.0.0-dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// String returns a single-line version suitable for CLI output.
func String() string {
	return fmt.Sprintf("%s (commit %s, built %s)", Version, GitCommit, BuildDate)
}
