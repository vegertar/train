package simnet

import "testing"

func TestGetLocation(t *testing.T) {
	t.Run("Checking all cities used by affinity.go", func(t *testing.T) {
		for _, name := range names {
			_, err := GetLocation(name)
			if err != nil {
				t.Error(err)
			}
		}
	})

	_, err := GetLocation("")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
