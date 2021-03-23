package short

import (
	"testing"
)

func TestMD5Hasher(t *testing.T) {
	h := NewMD5()

	got := h.Hash("test-text")

	if got != "cf0feea200efdea7d8580c7d4ef57ced" {
		t.Fatalf("expected %q, got %q", "cf0feea200efdea7d8580c7d4ef57ced", got)
	}
}
