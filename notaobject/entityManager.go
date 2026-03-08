package notaobject

import (
	"sync"
)

type Scene struct {
	Name     string
	entities map[string]*Entity
	mu       sync.RWMutex
}

func NewScene(name string) *Scene {
	return &Scene{
		Name:     name,
		entities: make(map[string]*Entity),
	}
}

// Add an entity (returns the entity for chaining)
func (s *Scene) Add(entity *Entity) *Entity {
	if entity == nil || entity.ID == "" {
		return entity
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entities[entity.ID] = entity
	return entity
}

// Remove an entity
func (s *Scene) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entities, id)
}

// Get an entity by ID
func (s *Scene) Get(id string) *Entity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entities[id]
}

// All returns all entities (for iteration)
func (s *Scene) All() []*Entity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entities := make([]*Entity, 0, len(s.entities))
	for _, e := range s.entities {
		entities = append(entities, e)
	}
	return entities
}

// Active returns only active entities
func (s *Scene) Active() []*Entity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var active []*Entity
	for _, e := range s.entities {
		if e.Active {
			active = append(active, e)
		}
	}
	return active
}

// Visible returns only visible entities
func (s *Scene) Visible() []*Entity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var visible []*Entity
	for _, e := range s.entities {
		if e.Visible {
			visible = append(visible, e)
		}
	}
	return visible
}

// Clear removes all entities
func (s *Scene) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entities = make(map[string]*Entity)
}
