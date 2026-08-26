package build

import (
	"log/slog"
	"time"
)

// These variables are set via ldflags. When unset, Nomen is the frozen
// 1.0.0-alpha product version. Keep this string identical to
// backend/v1/domain.ProductVersion — cmd/build must not import that package.
var (
	version = ""
	commit  = ""
	date    = ""
)

// dateTime is the parsed version of [date]
var dateTime time.Time

// init prevents race conditions when accessing dateTime and version.
func init() {
	var err error
	dateTime, err = time.Parse(time.RFC3339, date)
	if err != nil {
		slog.Warn("could not parse build date, using current time instead", "err", err)
		dateTime = time.Now()
		date = dateTime.Format(time.RFC3339)
	}
	if version == "" || version == "dev" || version == "worktree" {
		version = "1.0.0-alpha"
	}
}

// Version returns the current build version of Nomen
func Version() string {
	return version
}

// Commit returns the git commit hash of the current build of Nomen
func Commit() string {
	return commit
}

// Date returns the build date of the current build of Nomen
func Date() time.Time {
	return dateTime
}
