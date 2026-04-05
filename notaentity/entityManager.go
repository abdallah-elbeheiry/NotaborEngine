package notaentity

import (
	"NotaborEngine/notamath"
	"NotaborEngine/notatomic"

	"github.com/viterin/vek/vek32"
)

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

	entities    notatomic.Pointer[[]*Entity]
	idToIndex   map[string]int
	freeIndices []int
}

func NewEntityManager() *EntityManager {
	em := &EntityManager{
		idToIndex: make(map[string]int),
	}
	initialEntities := make([]*Entity, 0, 32)
	em.entities.Set(&initialEntities)
	return em
}

func (em *EntityManager) CreateEntity(id, name string) *Entity {
	var index int

	if len(em.freeIndices) > 0 {
		index = em.freeIndices[len(em.freeIndices)-1]
		em.freeIndices = em.freeIndices[:len(em.freeIndices)-1]
	} else {
		index = len(em.pos) / 2
		em.pos = append(em.pos, 0, 0)
		em.move = append(em.move, 0, 0)
		em.rot = append(em.rot, 0)
		em.rotD = append(em.rotD, 0)
		em.scale = append(em.scale, 1, 1)
		em.scaleD = append(em.scaleD, 1, 1)
	}

	e := newEntity(id, name, index, em)
	em.idToIndex[id] = index

	for {
		oldPtr := em.entities.Get()
		newSlice := append(*oldPtr, e)
		if em.entities.CompareAndSwap(oldPtr, &newSlice) {
			break
		}
	}

	return e
}

func (em *EntityManager) SubmitMove(index int, delta notamath.Vec2) {
	i := index * 2
	em.move[i] += delta.X
	em.move[i+1] += delta.Y
	em.dirtyMove.SetIfFalse(true)
}

func (em *EntityManager) SubmitRotation(index int, rad float32) {
	em.rotD[index] += rad
	em.dirtyRot.SetIfFalse(true)
}

func (em *EntityManager) SubmitScale(index int, factor notamath.Vec2) {
	i := index * 2
	em.scaleD[i] *= factor.X
	em.scaleD[i+1] *= factor.Y
	em.dirtyScale.SetIfFalse(true)
}

func (em *EntityManager) Flush() {
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
	em.syncColliders()
}

func (em *EntityManager) GetPosition(index int) notamath.Vec2 {
	i := index * 2
	return notamath.Vec2{X: em.pos[i], Y: em.pos[i+1]}
}

func (em *EntityManager) GetScale(index int) notamath.Vec2 {
	i := index * 2
	return notamath.Vec2{X: em.scale[i], Y: em.scale[i+1]}
}

func (em *EntityManager) GetRotation(index int) float32 {
	return em.rot[index]
}

func (em *EntityManager) GetEntities() []*Entity {
	return *em.entities.Get()
}

func (em *EntityManager) GetEntity(id string) *Entity {
	idx, exists := em.idToIndex[id]
	if !exists {
		return nil
	}
	return em.GetEntities()[idx]
}

func (em *EntityManager) Remove(id string) {
	idx, exists := em.idToIndex[id]
	if !exists {
		return
	}

	em.freeIndices = append(em.freeIndices, idx)
	delete(em.idToIndex, id)

	for {
		oldPtr := em.entities.Get()
		oldSlice := *oldPtr
		newSlice := make([]*Entity, 0, len(oldSlice)-1)
		for _, e := range oldSlice {
			if e.ID != id {
				newSlice = append(newSlice, e)
			}
		}
		if em.entities.CompareAndSwap(oldPtr, &newSlice) {
			break
		}
	}
}

func (em *EntityManager) syncColliders() {
	entities := *em.entities.Get()
	for _, e := range entities {
		e.updateCollider()
	}
}
