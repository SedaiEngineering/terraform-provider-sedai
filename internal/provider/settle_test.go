package provider

import "time"

// Tests run the post-switch wait and post-save check with millisecond gaps.
func init() {
	aggressionModeSettle = settleTiming{
		interval: time.Millisecond, timeout: 2 * time.Second, reads: 3,
		verifyInterval: time.Millisecond, verifyReads: 2,
	}
}
