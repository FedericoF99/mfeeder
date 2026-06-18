package shutdown

import (
	"context"
	"sync"
)

type Manager struct {
	Cancel context.CancelFunc
	once   sync.Once
}

func (m *Manager) Shutdown() {
	m.once.Do(func() {
		m.Cancel()
	})
}
