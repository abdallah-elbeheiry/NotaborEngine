package notaobject

import (
	"NotaborEngine/notatomic"
)

type Scene struct {
	Name string
	// We wrap the entire map in an atomic pointer
	entities notatomic.Pointer[map[string]*Entity]
}

func NewScene(name string) *Scene {
	s := &Scene{
		Name: name,
	}
	initialMap := make(map[string]*Entity)
	s.entities.Set(&initialMap)
	return s
}

// Add an entity
func (s *Scene) Add(entity *Entity) *Entity {
	if entity == nil || entity.ID == "" {
		return entity
	}

	for {
		oldMapPtr := s.entities.Get()
		// Shallow copy the map
		newMap := make(map[string]*Entity, len(*oldMapPtr)+1)
		for k, v := range *oldMapPtr {
			newMap[k] = v
		}
		newMap[entity.ID] = entity

		if s.entities.CompareAndSwap(oldMapPtr, &newMap) {
			return entity
		}
	}
}

// Remove an entity
func (s *Scene) Remove(id string) {
	for {
		oldMapPtr := s.entities.Get()
		if _, exists := (*oldMapPtr)[id]; !exists {
			return // Nothing to do
		}

		newMap := make(map[string]*Entity, len(*oldMapPtr))
		for k, v := range *oldMapPtr {
			if k != id {
				newMap[k] = v
			}
		}

		if s.entities.CompareAndSwap(oldMapPtr, &newMap) {
			return
		}
	}
}

// Get an entity by ID
func (s *Scene) Get(id string) *Entity {
	// Snapshot the pointer and read from the map snapshot
	m := *s.entities.Get()
	return m[id]
}

// All returns all entities
func (s *Scene) All() []*Entity {
	m := *s.entities.Get()
	entities := make([]*Entity, 0, len(m))
	for _, e := range m {
		entities = append(entities, e)
	}
	return entities
}

// Active returns only active entities
func (s *Scene) Active() []*Entity {
	m := *s.entities.Get()
	var active []*Entity
	for _, e := range m {
		if e.Active.Get() {
			active = append(active, e)
		}
	}
	return active
}

// Clear removes all entities
func (s *Scene) Clear() {
	emptyMap := make(map[string]*Entity)
	s.entities.Set(&emptyMap)
}
