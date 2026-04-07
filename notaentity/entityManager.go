package notaentity

import (
	"NotaborEngine/notacollision"
	"NotaborEngine/notamath"
	"NotaborEngine/notatomic"

	"github.com/viterin/vek/vek32"
)

// CollisionGroup is a set of entities that can collide with each other.
// Entities that don't share the same collision group will never collide.
type CollisionGroup struct {
	Name     string
	Entities []int
}

// collisionTable is a SuperHashMap (HashMapMap) which contains
// the collision status of all entity pairs.
type collisionTable struct {
	pairs map[int]map[int]notamath.Vec2
}

// transformData holds all transform arrays in an immutable wrapper
type transformData struct {
	pos   []float32
	rot   []float32
	scale []float32

	move   []float32
	rotD   []float32
	scaleD []float32
}

// indexMapping holds the ID-to-index mapping
type indexMapping struct {
	idToIndex   map[string]int
	freeIndices []int
}

// collisionGroupData holds all collision groups
type collisionGroupData struct {
	groups map[string]*CollisionGroup
}

// EntityManager manages entity transforms, batching of transform updates,
// collision groups, and collision detection results.
//
// Transform updates are submitted first and applied in bulk when flushing.
// Collision results are recalculated every frame after flushing colliders.
type EntityManager struct {
	transforms notatomic.Pointer[transformData]
	mapping    notatomic.Pointer[indexMapping]

	dirtyMove  notatomic.Bool
	dirtyRot   notatomic.Bool
	dirtyScale notatomic.Bool

	entities notatomic.Pointer[[]*Entity]

	collisionGroupsData notatomic.Pointer[collisionGroupData]
	collisionResults    notatomic.Pointer[collisionTable]
}

// NewEntityManager initializes the manager.
// This is called automatically via creating the engine,
// there is usually no need to call it manually.
func NewEntityManager() *EntityManager {

	em := &EntityManager{}

	initialTransforms := &transformData{
		pos:    make([]float32, 0, 64),
		rot:    make([]float32, 0, 32),
		scale:  make([]float32, 0, 64),
		move:   make([]float32, 0, 64),
		rotD:   make([]float32, 0, 32),
		scaleD: make([]float32, 0, 64),
	}
	em.transforms.Set(initialTransforms)

	initialMapping := &indexMapping{
		idToIndex:   make(map[string]int),
		freeIndices: make([]int, 0),
	}
	em.mapping.Set(initialMapping)

	initialEntities := make([]*Entity, 32)
	em.entities.Set(&initialEntities)

	initialGroups := &collisionGroupData{
		groups: make(map[string]*CollisionGroup),
	}
	em.collisionGroupsData.Set(initialGroups)

	initialTable := &collisionTable{
		pairs: make(map[int]map[int]notamath.Vec2),
	}
	em.collisionResults.Set(initialTable)

	return em
}

// CreateEntity creates a new entity.
// The given ID will be needed to access the entity from the manager,
// so choose a memorable ID.
func (em *EntityManager) CreateEntity(id string) *Entity {

	var index int

	for {
		oldMapping := em.mapping.Get()

		var newMapping *indexMapping
		if len(oldMapping.freeIndices) > 0 {
			// Reuse free index
			index = oldMapping.freeIndices[len(oldMapping.freeIndices)-1]
			newMapping = &indexMapping{
				idToIndex:   copyMap(oldMapping.idToIndex),
				freeIndices: append([]int(nil), oldMapping.freeIndices[:len(oldMapping.freeIndices)-1]...),
			}
		} else {
			// Allocate new index
			oldTransforms := em.transforms.Get()
			index = len(oldTransforms.rot)

			newMapping = &indexMapping{
				idToIndex:   copyMap(oldMapping.idToIndex),
				freeIndices: append([]int(nil), oldMapping.freeIndices...),
			}
		}

		newMapping.idToIndex[id] = index

		if em.mapping.CompareAndSwap(oldMapping, newMapping) {
			break
		}
	}

	for {
		oldTransforms := em.transforms.Get()

		if index < len(oldTransforms.rot) {
			// Index already exists, no expansion needed
			break
		}

		// Need to expand
		newTransforms := &transformData{
			pos:    append(append([]float32(nil), oldTransforms.pos...), 0, 0),
			rot:    append(append([]float32(nil), oldTransforms.rot...), 0),
			scale:  append(append([]float32(nil), oldTransforms.scale...), 1, 1),
			move:   append(append([]float32(nil), oldTransforms.move...), 0, 0),
			rotD:   append(append([]float32(nil), oldTransforms.rotD...), 0),
			scaleD: append(append([]float32(nil), oldTransforms.scaleD...), 1, 1),
		}

		if em.transforms.CompareAndSwap(oldTransforms, newTransforms) {
			break
		}
	}

	// Create entity
	e := newEntity(id, index, em)

	// CAS loop to add entity to entities slice
	for {
		oldPtr := em.entities.Get()
		oldSlice := *oldPtr

		newSlice := oldSlice

		if index >= len(newSlice) {
			grow := make([]*Entity, index+32)
			copy(grow, newSlice)
			newSlice = grow
		} else {
			newSlice = append([]*Entity(nil), oldSlice...)
		}

		newSlice[index] = e

		if em.entities.CompareAndSwap(oldPtr, &newSlice) {
			break
		}
	}

	return e
}

// submitMove queues up a movement request to the manager.
// Actual movement happens when the entity manager is flushed.
//
// This batching approach allows for bulk SIMD operations
// which reduces CPU cost compared to per-entity updates.
//
// Submitting is recommended inside a fast loop (≈60Hz+).
func (em *EntityManager) submitMove(index int, delta notamath.Vec2) {

	for {
		oldTransforms := em.transforms.Get()

		i := index * 2
		if i+1 >= len(oldTransforms.move) {
			return
		}

		newTransforms := &transformData{
			pos:    oldTransforms.pos,
			rot:    oldTransforms.rot,
			scale:  oldTransforms.scale,
			move:   append([]float32(nil), oldTransforms.move...),
			rotD:   oldTransforms.rotD,
			scaleD: oldTransforms.scaleD,
		}

		newTransforms.move[i] += delta.X
		newTransforms.move[i+1] += delta.Y

		if em.transforms.CompareAndSwap(oldTransforms, newTransforms) {
			em.dirtyMove.SetIfFalse(true)
			break
		}
	}
}

// submitRotation queues up a rotation request to the manager.
// Rotation uses radians, not degrees.
//
// Actual rotation happens when the manager is flushed.
//
// Submitting is recommended inside a fast loop (120Hz+).
func (em *EntityManager) submitRotation(index int, rad float32) {

	for {
		oldTransforms := em.transforms.Get()

		if index >= len(oldTransforms.rotD) {
			return
		}

		newTransforms := &transformData{
			pos:    oldTransforms.pos,
			rot:    oldTransforms.rot,
			scale:  oldTransforms.scale,
			move:   oldTransforms.move,
			rotD:   append([]float32(nil), oldTransforms.rotD...),
			scaleD: oldTransforms.scaleD,
		}

		newTransforms.rotD[index] += rad

		if em.transforms.CompareAndSwap(oldTransforms, newTransforms) {
			em.dirtyRot.SetIfFalse(true)
			break
		}
	}
}

// submitScale queues up a scale request to the manager.
//
// Scaling is multiplicative, not additive.
// Example:
// scale × 1.1 then × 0.5 results in 0.55.
//
// Actual scaling happens when the manager is flushed.
//
// Submitting is recommended inside a fast loop (≈60Hz+).
func (em *EntityManager) submitScale(index int, factor notamath.Vec2) {

	for {
		oldTransforms := em.transforms.Get()

		i := index * 2
		if i+1 >= len(oldTransforms.scaleD) {
			return
		}

		newTransforms := &transformData{
			pos:    oldTransforms.pos,
			rot:    oldTransforms.rot,
			scale:  oldTransforms.scale,
			move:   oldTransforms.move,
			rotD:   oldTransforms.rotD,
			scaleD: append([]float32(nil), oldTransforms.scaleD...),
		}

		newTransforms.scaleD[i] *= factor.X
		newTransforms.scaleD[i+1] *= factor.Y

		if em.transforms.CompareAndSwap(oldTransforms, newTransforms) {
			em.dirtyScale.SetIfFalse(true)
			break
		}
	}
}

// Flush updates all entity transforms and synchronizes colliders.
//
// Typically called once per frame before running collision detection or drawing.
func (em *EntityManager) Flush() {
	em.flushEntities()
	em.flushColliders()
}

// flushEntities applies all submitted movement, rotation,
// and scaling updates to entity transforms.
//
// If flushing occurs slower than submission,
// multiple queued submissions may accumulate.
func (em *EntityManager) flushEntities() {

	if em.dirtyMove.Get() {
		for {
			oldTransforms := em.transforms.Get()

			newPos := append([]float32(nil), oldTransforms.pos...)
			newMove := make([]float32, len(oldTransforms.move))

			vek32.Add_Inplace(newPos, oldTransforms.move)

			newTransforms := &transformData{
				pos:    newPos,
				rot:    oldTransforms.rot,
				scale:  oldTransforms.scale,
				move:   newMove,
				rotD:   oldTransforms.rotD,
				scaleD: oldTransforms.scaleD,
			}

			if em.transforms.CompareAndSwap(oldTransforms, newTransforms) {
				em.dirtyMove.Set(false)
				break
			}
		}
	}

	if em.dirtyRot.Get() {
		for {
			oldTransforms := em.transforms.Get()

			newRot := append([]float32(nil), oldTransforms.rot...)
			newRotD := make([]float32, len(oldTransforms.rotD))

			vek32.Add_Inplace(newRot, oldTransforms.rotD)

			newTransforms := &transformData{
				pos:    oldTransforms.pos,
				rot:    newRot,
				scale:  oldTransforms.scale,
				move:   oldTransforms.move,
				rotD:   newRotD,
				scaleD: oldTransforms.scaleD,
			}

			if em.transforms.CompareAndSwap(oldTransforms, newTransforms) {
				em.dirtyRot.Set(false)
				break
			}
		}
	}

	if em.dirtyScale.Get() {
		for {
			oldTransforms := em.transforms.Get()

			newScale := append([]float32(nil), oldTransforms.scale...)
			newScaleD := make([]float32, len(oldTransforms.scaleD))
			vek32.Ones_Into(newScaleD, len(newScaleD))

			vek32.Mul_Inplace(newScale, oldTransforms.scaleD)

			newTransforms := &transformData{
				pos:    oldTransforms.pos,
				rot:    oldTransforms.rot,
				scale:  newScale,
				move:   oldTransforms.move,
				rotD:   oldTransforms.rotD,
				scaleD: newScaleD,
			}

			if em.transforms.CompareAndSwap(oldTransforms, newTransforms) {
				em.dirtyScale.Set(false)
				break
			}
		}
	}
}

// flushColliders updates every collider's transform so that
// it matches its owning entity, then clears previous collision results.
func (em *EntityManager) flushColliders() {

	em.syncColliders()

	empty := &collisionTable{
		pairs: make(map[int]map[int]notamath.Vec2),
	}

	em.collisionResults.Set(empty)
}

func (em *EntityManager) GetPosition(id string) notamath.Vec2 {
	mapping := em.mapping.Get()
	index, ok := mapping.idToIndex[id]
	if !ok {
		return notamath.Vec2{}
	}
	return em.getPositionIndex(index)
}

func (em *EntityManager) GetScale(id string) notamath.Vec2 {
	mapping := em.mapping.Get()
	index, ok := mapping.idToIndex[id]
	if !ok {
		return notamath.Vec2{}
	}
	return em.getScaleIndex(index)
}

func (em *EntityManager) GetRotation(id string) float32 {
	mapping := em.mapping.Get()
	index, ok := mapping.idToIndex[id]
	if !ok {
		return 0
	}
	return em.getRotationIndex(index)
}

func (em *EntityManager) getPositionIndex(index int) notamath.Vec2 {
	transforms := em.transforms.Get()
	i := index * 2

	if i+1 >= len(transforms.pos) {
		return notamath.Vec2{}
	}

	return notamath.Vec2{
		X: transforms.pos[i],
		Y: transforms.pos[i+1],
	}
}

func (em *EntityManager) getScaleIndex(index int) notamath.Vec2 {
	transforms := em.transforms.Get()
	i := index * 2

	if i+1 >= len(transforms.scale) {
		return notamath.Vec2{}
	}

	return notamath.Vec2{
		X: transforms.scale[i],
		Y: transforms.scale[i+1],
	}
}

func (em *EntityManager) getRotationIndex(index int) float32 {
	transforms := em.transforms.Get()
	if index >= len(transforms.rot) {
		return 0
	}
	return transforms.rot[index]
}

// GetEntities returns a slice containing all entities in the manager.
// Useful for iteration across all entities.
func (em *EntityManager) GetEntities() []*Entity {
	return *em.entities.Get()
}

// GetEntity returns the entity with the given ID.
// Returns nil if the entity does not exist.
func (em *EntityManager) GetEntity(id string) *Entity {

	mapping := em.mapping.Get()
	idx, ok := mapping.idToIndex[id]
	if !ok {
		return nil
	}

	entities := *em.entities.Get()

	if idx >= len(entities) {
		return nil
	}

	return entities[idx]
}

// Remove removes the entity with the given ID from the manager.
// Its index becomes available for reuse by future entities.
func (em *EntityManager) Remove(id string) {

	// CAS loop to update mapping
	var idx int
	for {
		oldMapping := em.mapping.Get()

		var ok bool
		idx, ok = oldMapping.idToIndex[id]
		if !ok {
			return
		}

		newMapping := &indexMapping{
			idToIndex:   copyMap(oldMapping.idToIndex),
			freeIndices: append(append([]int(nil), oldMapping.freeIndices...), idx),
		}
		delete(newMapping.idToIndex, id)

		if em.mapping.CompareAndSwap(oldMapping, newMapping) {
			break
		}
	}

	for {
		oldPtr := em.entities.Get()
		oldSlice := *oldPtr

		newSlice := make([]*Entity, len(oldSlice))
		copy(newSlice, oldSlice)

		if idx < len(newSlice) {
			newSlice[idx] = nil
		}

		if em.entities.CompareAndSwap(oldPtr, &newSlice) {
			break
		}
	}

	for {
		oldGroupData := em.collisionGroupsData.Get()

		newGroupData := &collisionGroupData{
			groups: make(map[string]*CollisionGroup),
		}

		for name, g := range oldGroupData.groups {
			filter := make([]int, 0, len(g.Entities))

			for _, v := range g.Entities {
				if v != idx {
					filter = append(filter, v)
				}
			}

			newGroupData.groups[name] = &CollisionGroup{
				Name:     name,
				Entities: filter,
			}
		}

		if em.collisionGroupsData.CompareAndSwap(oldGroupData, newGroupData) {
			break
		}
	}
}

// AddToCollisionGroup adds the given entity to a collision group.
// If the group does not exist, it will be created.
//
// Entities within the same group can collide.
// Entities in different groups will never collide.
func (em *EntityManager) AddToCollisionGroup(group string, e *Entity) {

	for {
		oldGroupData := em.collisionGroupsData.Get()

		newGroupData := &collisionGroupData{
			groups: make(map[string]*CollisionGroup),
		}

		// Copy all existing groups
		for name, g := range oldGroupData.groups {
			newGroupData.groups[name] = &CollisionGroup{
				Name:     name,
				Entities: append([]int(nil), g.Entities...),
			}
		}

		// Add to target group
		if existingGroup, ok := newGroupData.groups[group]; ok {
			existingGroup.Entities = append(existingGroup.Entities, e.index)
		} else {
			newGroupData.groups[group] = &CollisionGroup{
				Name:     group,
				Entities: []int{e.index},
			}
		}

		if em.collisionGroupsData.CompareAndSwap(oldGroupData, newGroupData) {
			break
		}
	}
}

// SolveGroupCollision computes collisions between all entities
// inside the specified collision group.
//
// Must be called before querying collision results,
// since flushing clears previous collision data.
func (em *EntityManager) SolveGroupCollision(id string) {

	groupData := em.collisionGroupsData.Get()
	g, ok := groupData.groups[id]
	if !ok || g == nil {
		return
	}

	entities := *em.entities.Get()

	// Build new collision table
	newPairs := make(map[int]map[int]notamath.Vec2)

	for a := 0; a < len(g.Entities); a++ {

		i := g.Entities[a]

		e1 := entities[i]
		if e1 == nil {
			continue
		}

		c1 := e1.Collider.Get()
		if c1 == nil {
			continue
		}

		for b := a + 1; b < len(g.Entities); b++ {

			j := g.Entities[b]

			e2 := entities[j]
			if e2 == nil {
				continue
			}

			c2 := e2.Collider.Get()
			if c2 == nil {
				continue
			}

			_, mtv := notacollision.Intersects(*c1, *c2)
			if mtv != (notamath.Vec2{}) {
				if newPairs[i] == nil {
					newPairs[i] = make(map[int]notamath.Vec2)
				}
				if newPairs[j] == nil {
					newPairs[j] = make(map[int]notamath.Vec2)
				}
				newPairs[i][j] = mtv
				newPairs[j][i] = mtv.Neg() // inverse MTV
			}
		}
	}

	for {
		oldTable := em.collisionResults.Get()

		mergedPairs := make(map[int]map[int]notamath.Vec2)

		// Copy old pairs
		for k, v := range oldTable.pairs {
			mergedPairs[k] = make(map[int]notamath.Vec2)
			for k2, v2 := range v {
				mergedPairs[k][k2] = v2
			}
		}

		// Add new pairs
		for k, v := range newPairs {
			if mergedPairs[k] == nil {
				mergedPairs[k] = make(map[int]notamath.Vec2)
			}
			for k2, v2 := range v {
				mergedPairs[k][k2] = v2
			}
		}

		newTable := &collisionTable{
			pairs: mergedPairs,
		}

		if em.collisionResults.CompareAndSwap(oldTable, newTable) {
			break
		}
	}
}

func (em *EntityManager) syncColliders() {

	entities := *em.entities.Get()

	for _, e := range entities {

		if e == nil {
			continue
		}

		e.updateCollider()
	}
}

// Collides checks whether two entities are currently colliding.
func (em *EntityManager) Collides(a, b *Entity) bool {
	collides, _ := em.CollidesMTV(a, b)
	return collides
}

// GetMTV returns the minimum translation vector (MTV) between two entities.
// The MTV is the vector needed to move one entity to the other's position.
//
// If the entities are not colliding, the MTV is (0, 0).
func (em *EntityManager) GetMTV(a, b *Entity) notamath.Vec2 {
	_, mtv := em.CollidesMTV(a, b)
	return mtv
}

// CollidesMTV checks whether two entities are currently colliding,
// and returns the minimum translation vector (MTV) between them.
//
// If the entities are not colliding, the MTV is (0, 0).
func (em *EntityManager) CollidesMTV(a, b *Entity) (bool, notamath.Vec2) {
	table := em.collisionResults.Get()
	row := table.pairs[a.index]
	if row == nil {
		return false, notamath.Vec2{}
	}

	mtv, ok := row[b.index]
	return ok, mtv
}

// copyMap creates a copy of the given map
func copyMap(m map[string]int) map[string]int {
	result := make(map[string]int, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
