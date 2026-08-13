package access

import "testing"

func TestClassificationClearance(t *testing.T) {
	if CanRead(Internal, Confidential) { t.Fatal("internal clearance must not read confidential content") }
	if !CanRead(Restricted, Confidential) { t.Fatal("restricted clearance should read confidential content") }
}
