package sqljs

import (
	"testing"
)

func TestDb(t *testing.T) {
	db := New()

	if err := db.Close(); err != nil {
		t.Fatalf("Error closing DB: %s", err)
	}
}
