package system

import "github.com/ajkula/shmup/types"

type Rect struct {
	X, Y, Width, Height float64
}

type Quadtree struct {
	Boundary                                   Rect
	Capacity                                   int
	Entities                                   []types.GameEntity
	NorthWest, NorthEast, SouthWest, SouthEast *Quadtree
	Subdivided                                 bool
}

func NewQuadtree(boundary Rect, capacity int) *Quadtree {
	return &Quadtree{
		Boundary: boundary,
		Capacity: capacity,
		Entities: make([]types.GameEntity, 0, capacity),
	}
}

func (qt *Quadtree) Insert(entity types.GameEntity) bool {
	pos := entity.GetPosition()
	if !qt.Boundary.Contains(pos) {
		return false
	}

	if len(qt.Entities) < qt.Capacity && !qt.Subdivided {
		qt.Entities = append(qt.Entities, entity)
		return true
	}

	if !qt.Subdivided {
		qt.Subdivide()
	}

	return qt.NorthWest.Insert(entity) ||
		qt.NorthEast.Insert(entity) ||
		qt.SouthWest.Insert(entity) ||
		qt.SouthEast.Insert(entity)
}

func (qt *Quadtree) Remove(entity types.Entity) bool {
	pos := entity.GetPosition()
	if !qt.Boundary.Contains(pos) {
		return false
	}

	for i, e := range qt.Entities {
		if e == entity {
			qt.Entities[i] = qt.Entities[len(qt.Entities)-1]
			qt.Entities = qt.Entities[:len(qt.Entities)-1]
			return true
		}
	}

	if qt.Subdivided {
		return qt.NorthWest.Remove(entity) ||
			qt.NorthEast.Remove(entity) ||
			qt.SouthWest.Remove(entity) ||
			qt.SouthEast.Remove(entity)
	}

	return false
}

// Subdivide splits the quadtree into 4 smaller quadrants
func (qt *Quadtree) Subdivide() {
	x := qt.Boundary.X
	y := qt.Boundary.Y
	w := qt.Boundary.Width / 2
	h := qt.Boundary.Height / 2

	qt.NorthWest = NewQuadtree(Rect{x, y, w, h}, qt.Capacity)
	qt.NorthEast = NewQuadtree(Rect{x + w, y, w, h}, qt.Capacity)
	qt.SouthWest = NewQuadtree(Rect{x, y + h, w, h}, qt.Capacity)
	qt.SouthEast = NewQuadtree(Rect{x + w, y + h, w, h}, qt.Capacity)

	qt.Subdivided = true
	// Reinsert all entities
	for _, entity := range qt.Entities {
		qt.NorthWest.Insert(entity)
		qt.NorthEast.Insert(entity)
		qt.SouthWest.Insert(entity)
		qt.SouthEast.Insert(entity)
	}
	qt.Entities = nil
}

// Query returns all entities that are within the range of the given entity
func (qt *Quadtree) Query(rg Rect) []types.GameEntity {
	var found []types.GameEntity

	if !qt.Boundary.Intersects(rg) {
		return found
	}

	for _, entity := range qt.Entities {
		if rg.Contains(entity.GetPosition()) {
			found = append(found, entity)
		}
	}

	if qt.Subdivided {
		found = append(found, qt.NorthWest.Query(rg)...)
		found = append(found, qt.NorthEast.Query(rg)...)
		found = append(found, qt.SouthWest.Query(rg)...)
		found = append(found, qt.SouthEast.Query(rg)...)
	}

	return found
}

func (r Rect) Contains(point types.Vector2D) bool {
	return point.X >= r.X && point.X < r.X+r.Width &&
		point.Y >= r.Y && point.Y < r.Y+r.Height
}

func (r Rect) Intersects(other Rect) bool {
	return r.X < other.X+other.Width &&
		r.X+r.Width > other.X &&
		r.Y < other.Y+other.Height &&
		r.Y+r.Height > other.Y
}

func (qt *Quadtree) Clear() {
	var stack []*Quadtree
	stack = append(stack, qt)

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		node.Entities = node.Entities[:0]

		if node.Subdivided {
			stack = append(stack, node.NorthWest, node.NorthEast, node.SouthWest, node.SouthEast)
			node.NorthWest = nil
			node.NorthEast = nil
			node.SouthWest = nil
			node.SouthEast = nil
			node.Subdivided = false
		}
	}
}

// GetAllEntities returns all entities in the quadtree
func (qt *Quadtree) GetAllEntities() []types.GameEntity {
	var entities []types.GameEntity
	entities = append(entities, qt.Entities...)

	if qt.Subdivided {
		entities = append(entities, qt.NorthWest.GetAllEntities()...)
		entities = append(entities, qt.NorthEast.GetAllEntities()...)
		entities = append(entities, qt.SouthWest.GetAllEntities()...)
		entities = append(entities, qt.SouthEast.GetAllEntities()...)
	}

	return entities
}

func (qt *Quadtree) DetectCollisions(callback func(e1, e2 types.Entity)) {
	var detectCollisionsRecursive func(*Quadtree)
	detectCollisionsRecursive = func(node *Quadtree) {
		// Check collisions within this node
		for i := 0; i < len(node.Entities); i++ {
			if !node.Entities[i].IsAlive() {
				continue
			}
			for j := i + 1; j < len(node.Entities); j++ {
				if !node.Entities[j].IsAlive() {
					continue
				}
				if checkCollision(node.Entities[i], node.Entities[j]) {
					callback(node.Entities[i], node.Entities[j])
				}
			}
		}

		if node.Subdivided {
			detectCollisionsRecursive(node.NorthWest)
			detectCollisionsRecursive(node.NorthEast)
			detectCollisionsRecursive(node.SouthWest)
			detectCollisionsRecursive(node.SouthEast)
		}
	}

	detectCollisionsRecursive(qt)
}

func checkCollision(e1, e2 types.Entity) bool {
	x1, y1, w1, h1 := e1.GetCollisionBox()
	x2, y2, w2, h2 := e2.GetCollisionBox()
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}
