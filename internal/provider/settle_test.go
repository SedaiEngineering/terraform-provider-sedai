package provider

import "time"

// Tests run the post-switch wait with millisecond gaps.
func init() {
	aggressionModeSettle = settleTiming{interval: time.Millisecond, timeout: 2 * time.Second, reads: 3}
}
