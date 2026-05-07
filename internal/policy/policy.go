// Package policy determines which operational constraints are currently active,
// such as peak-hour quota burn rates.
package policy

import (
	"time"
)

// IsPeakHours reports whether t falls within the peak usage window for the
// given subscription type. During peak hours, quota limits burn faster for
// affected subscription types (pro and max plans).
// Currently, Anthropic does not employ peak-hours, so this feature is
// disabled for now.
func IsPeakHours(_ time.Time, _ string) bool {
	return false
}
