// Package core provides the in-memory stores for Connections and Workflows.
package core

import (
	"fmt"
	"sync"
)

// ConnectionStore is a thread-safe in-memory store for Connection objects.
type ConnectionStore struct {
	mu   sync.RWMutex
	data map[string]Connection
}

// NewConnectionStore returns an initialised ConnectionStore.
func NewConnectionStore() *ConnectionStore {
	return &ConnectionStore{data: make(map[string]Connection)}
}

// Add inserts a new Connection. Returns an error if the ID already exists.
func (s *ConnectionStore) Add(c Connection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[c.ID]; exists {
		return fmt.Errorf("connection %q already exists", c.ID)
	}
	s.data[c.ID] = c
	return nil
}

// Get retrieves a Connection by ID.
func (s *ConnectionStore) Get(id string) (Connection, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.data[id]
	return c, ok
}

// List returns all stored connections.
func (s *ConnectionStore) List() []Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Connection, 0, len(s.data))
	for _, c := range s.data {
		out = append(out, c)
	}
	return out
}

// Delete removes a Connection by ID. Returns an error if not found.
func (s *ConnectionStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[id]; !exists {
		return fmt.Errorf("connection %q not found", id)
	}
	delete(s.data, id)
	return nil
}

// ReplaceAll clears the store and repopulates it with conns. Used when
// importing a configuration bundle.
func (s *ConnectionStore) ReplaceAll(conns []Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]Connection, len(conns))
	for _, c := range conns {
		s.data[c.ID] = c
	}
}

// WorkflowStore is a thread-safe in-memory store for Workflow objects.
type WorkflowStore struct {
	mu   sync.RWMutex
	data map[string]Workflow
}

// NewWorkflowStore returns an initialised WorkflowStore.
func NewWorkflowStore() *WorkflowStore {
	return &WorkflowStore{data: make(map[string]Workflow)}
}

// Add inserts a new Workflow. Returns an error if the ID already exists.
func (s *WorkflowStore) Add(w Workflow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[w.ID]; exists {
		return fmt.Errorf("workflow %q already exists", w.ID)
	}
	s.data[w.ID] = w
	return nil
}

// Get retrieves a Workflow by ID.
func (s *WorkflowStore) Get(id string) (Workflow, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.data[id]
	return w, ok
}

// List returns all stored workflows.
func (s *WorkflowStore) List() []Workflow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Workflow, 0, len(s.data))
	for _, w := range s.data {
		out = append(out, w)
	}
	return out
}

// Delete removes a Workflow by ID. Returns an error if not found.
func (s *WorkflowStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[id]; !exists {
		return fmt.Errorf("workflow %q not found", id)
	}
	delete(s.data, id)
	return nil
}

// Update replaces an existing Workflow. Returns an error if the workflow does not exist.
func (s *WorkflowStore) Update(w Workflow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[w.ID]; !exists {
		return fmt.Errorf("workflow %q not found", w.ID)
	}
	s.data[w.ID] = w
	return nil
}

// ReplaceAll clears the store and repopulates it with flows. Used when
// importing a configuration bundle.
func (s *WorkflowStore) ReplaceAll(flows []Workflow) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]Workflow, len(flows))
	for _, w := range flows {
		s.data[w.ID] = w
	}
}
