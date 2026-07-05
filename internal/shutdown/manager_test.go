package shutdown

import "testing"

func TestShutdownCallsCancelOnce(t *testing.T) {
	calls := 0
	manager := Manager{
		Cancel: func() {
			calls++
		},
	}

	manager.Shutdown()
	manager.Shutdown()
	manager.Shutdown()

	if calls != 1 {
		t.Fatalf("expected cancel to be called once, got %d", calls)
	}
}
