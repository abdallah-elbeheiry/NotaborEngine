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
	Entities []int // entity indices
}

// collisionTable is a SuperHashMap (HashMapMap) which contains
// the collision status of all entity pairs.
type collisionTable struct {
	pairs map[int]map[int]bool
}

// EntityManager manages entity transforms, batching of transform updates,
// collision groups, and collision detection results.
//
// Transform updates are submitted first and applied in bulk when flushing.
// Collision results are recalculated every frame after flushing colliders.
type EntityManager struct {
	pos   []float32
	rot   []float32
	scale []float32

	move   []float32
	rotD   []float32
	scaleD []float32

	dirtyMove  notatomic.Bool
	dirtyRot   notatomic.Bool
	dirtyScale notatomic.Bool

	entities notatomic.Pointer[[]*Entity]

	collisionGroups  map[string]*CollisionGroup
	collisionResults notatomic.Pointer[collisionTable]

	idToIndex   map[string]int
	freeIndices []int
}

// NewEntityManager initializes the manager.
// This is called automatically via creating the engine,
// there is usually no need to call it manually.
func NewEntityManager() *EntityManager {

	em := &EntityManager{
		idToIndex:       make(map[string]int),
		collisionGroups: make(map[string]*CollisionGroup),
	}

	initialEntities := make([]*Entity, 32)
	em.entities.Set(&initialEntities)

	initialTable := &collisionTable{
		pairs: make(map[int]map[int]bool),
	}
	em.collisionResults.Set(initialTable)

	return em
}

// CreateEntity creates a new entity.
// The given ID will be needed to access the entity from the manager,
// so choose a memorable ID.
func (em *EntityManager) CreateEntity(id string) *Entity {

	var index int

	if len(em.freeIndices) > 0 {
		index = em.freeIndices[len(em.freeIndices)-1]
		em.freeIndices = em.freeIndices[:len(em.freeIndices)-1]
	} else {

		index = len(em.rot)

		em.pos = append(em.pos, 0, 0)
		em.move = append(em.move, 0, 0)

		em.rot = append(em.rot, 0)
		em.rotD = append(em.rotD, 0)

		em.scale = append(em.scale, 1, 1)
		em.scaleD = append(em.scaleD, 1, 1)
	}

	e := newEntity(id, index, em)
	em.idToIndex[id] = index

	for {

		oldPtr := em.entities.Get()
		oldSlice := *oldPtr

		newSlice := oldSlice

		if index >= len(newSlice) {
			grow := make([]*Entity, index+32)
			copy(grow, newSlice)
			newSlice = grow
		}

		newSlice[index] = e

		if em.entities.CompareAndSwap(oldPtr, &newSlice) {
			break
		}
	}

	return e
}

// SubmitMove queues up a movement request to the manager.
// Actual movement happens when the entity manager is flushed.
//
// This batching approach allows for bulk SIMD operations
// which reduces CPU cost compared to per-entity updates.
//
// Submitting is recommended inside a fast loop (≈60Hz+).
func (em *EntityManager) SubmitMove(index int, delta notamath.Vec2) {

	i := index * 2
	em.move[i] += delta.X
	em.move[i+1] += delta.Y

	em.dirtyMove.SetIfFalse(true)
}

// SubmitRotation queues up a rotation request to the manager.
// Rotation uses radians, not degrees.
//
// Actual rotation happens when the manager is flushed.
//
// Submitting is recommended inside a fast loop (120Hz+).
func (em *EntityManager) SubmitRotation(index int, rad float32) {

	em.rotD[index] += rad
	em.dirtyRot.SetIfFalse(true)
}

// SubmitScale queues up a scale request to the manager.
//
// Scaling is multiplicative, not additive.
// Example:
// scale × 1.1 then × 0.5 results in 0.55.
//
// Actual scaling happens when the manager is flushed.
//
// Submitting is recommended inside a fast loop (≈60Hz+).
func (em *EntityManager) SubmitScale(index int, factor notamath.Vec2) {

	i := index * 2

	em.scaleD[i] *= factor.X
	em.scaleD[i+1] *= factor.Y

	em.dirtyScale.SetIfFalse(true)
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

		vek32.Add_Inplace(em.pos, em.move)
		vek32.Zeros_Into(em.move, len(em.move))

		em.dirtyMove.Set(false)
	}

	if em.dirtyRot.Get() {

		vek32.Add_Inplace(em.rot, em.rotD)
		vek32.Zeros_Into(em.rotD, len(em.rotD))

		em.dirtyRot.Set(false)
	}

	if em.dirtyScale.Get() {

		vek32.Mul_Inplace(em.scale, em.scaleD)
		vek32.Ones_Into(em.scaleD, len(em.scaleD))

		em.dirtyScale.Set(false)
	}
}

// flushColliders updates every collider's transform so that
// it matches its owning entity, then clears previous collision results.
func (em *EntityManager) flushColliders() {

	em.syncColliders()

	empty := &collisionTable{
		pairs: make(map[int]map[int]bool),
	}

	em.collisionResults.Set(empty)
}

func (em *EntityManager) GetPosition(id string) notamath.Vec2 {
	return em.getPositionIndex(em.idToIndex[id])
}

func (em *EntityManager) GetScale(id string) notamath.Vec2 {
	return em.getScaleIndex(em.idToIndex[id])
}

func (em *EntityManager) GetRotation(id string) float32 {
	return em.getRotationIndex(em.idToIndex[id])
}

func (em *EntityManager) getPositionIndex(index int) notamath.Vec2 {

	i := index * 2

	return notamath.Vec2{
		X: em.pos[i],
		Y: em.pos[i+1],
	}
}

func (em *EntityManager) getScaleIndex(index int) notamath.Vec2 {

	i := index * 2

	return notamath.Vec2{
		X: em.scale[i],
		Y: em.scale[i+1],
	}
}

func (em *EntityManager) getRotationIndex(index int) float32 {
	return em.rot[index]
}

// GetEntities returns a slice containing all entities in the manager.
// Useful for iteration across all entities.
func (em *EntityManager) GetEntities() []*Entity {
	return *em.entities.Get()
}

// GetEntity returns the entity with the given ID.
// Returns nil if the entity does not exist.
func (em *EntityManager) GetEntity(id string) *Entity {

	idx, ok := em.idToIndex[id]
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

	idx, ok := em.idToIndex[id]
	if !ok {
		return
	}

	em.freeIndices = append(em.freeIndices, idx)
	delete(em.idToIndex, id)

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

	for _, g := range em.collisionGroups {

		filter := g.Entities[:0]

		for _, v := range g.Entities {

			if v != idx {
				filter = append(filter, v)
			}
		}

		g.Entities = filter
	}
}

// AddToCollisionGroup adds the given entity to a collision group.
// If the group does not exist, it will be created.
//
// Entities within the same group can collide.
// Entities in different groups will never collide.
func (em *EntityManager) AddToCollisionGroup(group string, e *Entity) {

	g, ok := em.collisionGroups[group]

	if !ok {
		g = &CollisionGroup{Name: group}
		em.collisionGroups[group] = g
	}

	g.Entities = append(g.Entities, e.index)
}

// SolveGroupCollision computes collisions between all entities
// inside the specified collision group.
//
// Must be called before querying collision results,
// since flushing clears previous collision data.
func (em *EntityManager) SolveGroupCollision(id string) {

	g := em.collisionGroups[id]
	if g == nil {
		return
	}

	entities := *em.entities.Get()
	table := em.collisionResults.Get()

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

			if notacollision.Intersects(*c1, *c2) {
				storeCollision(table, i, j)
			}
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

func storeCollision(table *collisionTable, a, b int) {

	if table.pairs[a] == nil {
		table.pairs[a] = make(map[int]bool)
	}

	if table.pairs[b] == nil {
		table.pairs[b] = make(map[int]bool)
	}

	table.pairs[a][b] = true
	table.pairs[b][a] = true
}

// Collides checks whether two entities are currently colliding.
func (em *EntityManager) Collides(a, b *Entity) bool {

	table := em.collisionResults.Get()

	row := table.pairs[a.index]
	if row == nil {
		return false
	}

	return row[b.index]
}
