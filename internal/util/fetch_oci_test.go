package util

import (
	"testing"
)

func TestFetchOCI_Invalid(t *testing.T) {
	if _, err := FetchOCI("oci://bad"); err == nil {
		t.Error("expected error for invalid ref")
	}
}
