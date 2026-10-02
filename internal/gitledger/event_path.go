package gitledger

// ValidateEventPath exposes the durable event-path grammar used by candidate
// construction so read-only namespace validation and writer validation cannot
// silently diverge.
func ValidateEventPath(path string) error {
	return validateEventPath(path)
}
