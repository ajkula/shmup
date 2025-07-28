package system

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ajkula/shmup/entity"
	"github.com/ajkula/shmup/mocks"
	"github.com/ajkula/shmup/types"
)

type MockEventManager struct{}

func (m *MockEventManager) Initialize() {}

func TestRectContains(t *testing.T) {
	rect := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	pointInside := types.Vector2D{X: 5, Y: 5}
	pointOutside := types.Vector2D{X: 15, Y: 15}

	if !rect.Contains(pointInside) {
		t.Errorf("Expected point %v to be inside rect %v", pointInside, rect)
	}

	if rect.Contains(pointOutside) {
		t.Errorf("Expected point %v to be outside rect %v", pointOutside, rect)
	}
}

func TestRectIntersects(t *testing.T) {
	rect1 := Rect{X: 0, Y: 0, Width: 10, Height: 10}
	rect2 := Rect{X: 5, Y: 5, Width: 10, Height: 10}
	rect3 := Rect{X: 15, Y: 15, Width: 10, Height: 10}

	if !rect1.Intersects(rect2) {
		t.Errorf("Expected rect %v to intersect with rect %v", rect1, rect2)
	}

	if rect1.Intersects(rect3) {
		t.Errorf("Expected rect %v to not intersect with rect %v", rect1, rect3)
	}
}

func TestQuadtreeClear(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 4)

	// Add some entities to the quadtree
	eventManager := &mocks.MockEventManager{}
	entity1 := mocks.NewMockEnemy(eventManager)
	entity1.SetPosition(types.Vector2D{X: 10, Y: 10})
	qt.Insert(entity1)

	entity2 := mocks.NewMockEnemy(eventManager)
	entity2.SetPosition(types.Vector2D{X: 20, Y: 20})
	qt.Insert(entity2)

	qt.Clear()

	if len(qt.Entities) != 0 {
		t.Errorf("Expected quadtree entities to be cleared, but found %d entities", len(qt.Entities))
	}

	if qt.Subdivided {
		t.Errorf("Expected quadtree to not be subdivided after clear")
	}
}

func TestQuadtreeQuery(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 4)

	eventManager := &mocks.MockEventManager{}
	entity1 := mocks.NewMockEnemy(eventManager)
	entity1.SetPosition(types.Vector2D{X: 10, Y: 10})
	entity2 := mocks.NewMockEnemy(eventManager)
	entity2.SetPosition(types.Vector2D{X: 20, Y: 20})
	entity3 := mocks.NewMockEnemy(eventManager)
	entity3.SetPosition(types.Vector2D{X: 110, Y: 110})

	qt.Insert(entity1)
	qt.Insert(entity2)
	qt.Insert(entity3)

	found := qt.Query(Rect{X: 0, Y: 0, Width: 50, Height: 50})

	if len(found) != 2 {
		t.Errorf("Expected to find 2 entities, but found %d", len(found))
	}

	for _, entity := range found {
		if entity == entity3 {
			t.Errorf("Expected entity %v to not be found in the query", entity3)
		}
	}
}

func TestQuadtreeCollision(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 4)

	eventManager := mocks.NewMockEventManager()
	entity1 := mocks.NewMockEnemy(eventManager)
	entity1.SetPosition(types.Vector2D{X: 10, Y: 10})
	entity2 := mocks.NewMockEnemy(eventManager)
	entity2.SetPosition(types.Vector2D{X: 15, Y: 15})
	entity3 := mocks.NewMockEnemy(eventManager)
	entity3.SetPosition(types.Vector2D{X: 80, Y: 80})

	qt.Insert(entity1)
	qt.Insert(entity2)
	qt.Insert(entity3)

	found := qt.Query(Rect{X: 0, Y: 0, Width: 30, Height: 30})

	if len(found) != 2 {
		t.Errorf("Expected to find 2 entities in collision area, but found %d", len(found))
	}

	for _, entity := range found {
		if entity == entity3 {
			t.Errorf("Expected entity %v to not be found in the collision area", entity3)
		}
	}
}

func TestQuadtreeRemove(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 4)
	eventManager := &mocks.MockEventManager{}

	entity1 := mocks.NewMockEnemy(eventManager)
	entity1.SetPosition(types.Vector2D{X: 10, Y: 10})
	entity2 := mocks.NewMockEnemy(eventManager)
	entity2.SetPosition(types.Vector2D{X: 20, Y: 20})

	qt.Insert(entity1)
	qt.Insert(entity2)

	if !qt.Remove(entity1) {
		t.Errorf("Failed to remove entity1 from quadtree")
	}

	found := qt.Query(Rect{X: 0, Y: 0, Width: 100, Height: 100})
	if len(found) != 1 || found[0] != entity2 {
		t.Errorf("Expected only entity2 to remain in quadtree after removal")
	}

	if qt.Remove(entity1) {
		t.Errorf("Removing already removed entity should return false")
	}
}

func TestQuadtreeGetAllEntities(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 2) // Capacity of 2 to force subdivision
	eventManager := &mocks.MockEventManager{}

	entity1 := mocks.NewMockEnemy(eventManager)
	entity1.SetPosition(types.Vector2D{X: 10, Y: 10})
	entity2 := mocks.NewMockEnemy(eventManager)
	entity2.SetPosition(types.Vector2D{X: 20, Y: 20})
	entity3 := mocks.NewMockEnemy(eventManager)
	entity3.SetPosition(types.Vector2D{X: 80, Y: 80})

	qt.Insert(entity1)
	qt.Insert(entity2)
	qt.Insert(entity3)

	allEntities := qt.GetAllEntities()
	if len(allEntities) != 3 {
		t.Errorf("Expected to get all 3 entities, but got %d", len(allEntities))
	}

	entitySet := make(map[types.GameEntity]bool)
	for _, entity := range allEntities {
		entitySet[entity] = true
	}

	if !entitySet[entity1] || !entitySet[entity2] || !entitySet[entity3] {
		t.Errorf("GetAllEntities did not return all inserted entities")
	}
}

func TestQuadtreeSubdivision(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 2) // Capacity of 2 to force subdivision
	eventManager := &mocks.MockEventManager{}

	entity1 := mocks.NewMockEnemy(eventManager)
	entity1.SetPosition(types.Vector2D{X: 10, Y: 10})
	entity2 := mocks.NewMockEnemy(eventManager)
	entity2.SetPosition(types.Vector2D{X: 20, Y: 20})
	entity3 := mocks.NewMockEnemy(eventManager)
	entity3.SetPosition(types.Vector2D{X: 80, Y: 80})

	qt.Insert(entity1)
	qt.Insert(entity2)

	if qt.Subdivided {
		t.Errorf("Quadtree should not be subdivided with only 2 entities")
	}

	qt.Insert(entity3)

	if !qt.Subdivided {
		t.Errorf("Quadtree should be subdivided after inserting 3rd entity")
	}

	if len(qt.Entities) != 0 {
		t.Errorf("Parent quadtree should have no entities after subdivision")
	}

	// Check if entities are in correct sub-quadrants
	nwEntities := qt.NorthWest.GetAllEntities()
	neEntities := qt.NorthEast.GetAllEntities()
	swEntities := qt.SouthWest.GetAllEntities()
	seEntities := qt.SouthEast.GetAllEntities()

	if len(nwEntities) != 2 || len(seEntities) != 1 || len(neEntities) != 0 || len(swEntities) != 0 {
		t.Errorf("Entities are not correctly distributed in sub-quadrants")
	}
}

func TestQuadtreeEdgeCases(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 4)
	eventManager := &mocks.MockEventManager{}

	// Test inserting entity on the edge
	edgeEntity := mocks.NewMockEnemy(eventManager)
	edgeEntity.SetPosition(types.Vector2D{X: 100, Y: 50})
	if qt.Insert(edgeEntity) {
		t.Errorf("Entity on the edge should not be inserted")
	}

	// Test inserting entity outside bounds
	outsideEntity := mocks.NewMockEnemy(eventManager)
	outsideEntity.SetPosition(types.Vector2D{X: 150, Y: 150})
	if qt.Insert(outsideEntity) {
		t.Errorf("Entity outside bounds should not be inserted")
	}

	// Test querying empty area
	emptyQuery := qt.Query(Rect{X: 200, Y: 200, Width: 50, Height: 50})
	if len(emptyQuery) != 0 {
		t.Errorf("Query on empty area should return no entities")
	}

	// Test removing non-existent entity
	nonExistentEntity := mocks.NewMockEnemy(eventManager)
	nonExistentEntity.SetPosition(types.Vector2D{X: 50, Y: 50})
	if qt.Remove(nonExistentEntity) {
		t.Errorf("Removing non-existent entity should return false")
	}
}

func TestDetectCollisions(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 4)
	eventManager := &mocks.MockEventManager{}

	player := entity.NewPlayer(types.Vector2D{X: 10, Y: 10}, eventManager)
	enemy := entity.NewEnemy(types.Vector2D{X: 20, Y: 20}, eventManager)
	bullet := entity.NewBullet(15, 15, false, eventManager)

	qt.Insert(player)
	qt.Insert(enemy)
	qt.Insert(bullet)

	collisions := make(map[string]bool)
	expectedCollisions := []string{
		"Enemy-Player",
		"Bullet-Player",
		"Bullet-Enemy",
	}

	qt.DetectCollisions(func(e1, e2 types.Entity) {
		collisionKey := getCollisionKey(e1, e2)
		t.Logf("Collision detected: %s", collisionKey)
		collisions[collisionKey] = true
	})

	for _, expected := range expectedCollisions {
		if !collisions[expected] {
			t.Errorf("Expected collision %s was not detected", expected)
		}
	}

	if len(collisions) != len(expectedCollisions) {
		t.Errorf("Expected %d unique collisions, but got %d", len(expectedCollisions), len(collisions))
	}
}

func getCollisionKey(e1, e2 types.Entity) string {
	type1 := reflect.TypeOf(e1).Elem().Name()
	type2 := reflect.TypeOf(e2).Elem().Name()
	if type1 > type2 {
		type1, type2 = type2, type1
	}
	return fmt.Sprintf("%s-%s", type1, type2)
}

func TestDetectCollisionsWithDeadEntities(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 4)
	eventManager := &mocks.MockEventManager{}

	player := entity.NewPlayer(types.Vector2D{X: 10, Y: 10}, eventManager)
	enemy := entity.NewEnemy(types.Vector2D{X: 20, Y: 20}, eventManager)
	enemy.TakeDamage(1000) // Kill enemy

	qt.Insert(player)
	qt.Insert(enemy)

	collisions := 0
	qt.DetectCollisions(func(e1, e2 types.Entity) {
		collisions++
	})

	if collisions != 0 {
		t.Errorf("Expected 0 collisions with dead entity, but got %d", collisions)
	}
}

func TestQuadtreeInsertBoundary(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 4)
	eventManager := &mocks.MockEventManager{}

	// Test inserting entity exactly on the boundary
	boundaryEntity := entity.NewEnemy(types.Vector2D{X: 100, Y: 100}, eventManager)

	if qt.Insert(boundaryEntity) {
		t.Errorf("Entity exactly on the boundary should not be inserted")
	}
}

func TestQuadtreeSubdivisionAndQuery(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 2)
	eventManager := &mocks.MockEventManager{}

	for i := 0; i < 10; i++ {
		enemy := entity.NewEnemy(types.Vector2D{X: float64(i * 10), Y: float64(i * 10)}, eventManager)
		qt.Insert(enemy)
	}

	queryRect := Rect{X: 25, Y: 25, Width: 50, Height: 50}
	results := qt.Query(queryRect)

	if len(results) == 0 {
		t.Errorf("Expected to find entities in query, but found none")
	}

	for _, entity := range results {
		pos := entity.GetPosition()
		if !queryRect.Contains(pos) {
			t.Errorf("Query returned entity outside of query rectangle: %v", pos)
		}
	}
}

func TestQuadtreeRemoveAndReinsert(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 4)
	eventManager := &mocks.MockEventManager{}

	enemy := entity.NewEnemy(types.Vector2D{X: 50, Y: 50}, eventManager)

	qt.Insert(enemy)

	if !qt.Remove(enemy) {
		t.Errorf("Failed to remove inserted entity")
	}

	if len(qt.GetAllEntities()) != 0 {
		t.Errorf("Quadtree should be empty after removing the only entity")
	}

	if !qt.Insert(enemy) {
		t.Errorf("Failed to reinsert entity")
	}

	if len(qt.GetAllEntities()) != 1 {
		t.Errorf("Quadtree should contain one entity after reinsertion")
	}
}

func TestQuadtreeClearSubdivided(t *testing.T) {
	qt := NewQuadtree(Rect{X: 0, Y: 0, Width: 100, Height: 100}, 2)
	eventManager := &mocks.MockEventManager{}

	for i := 0; i < 5; i++ {
		enemy := entity.NewEnemy(types.Vector2D{X: float64(i * 20), Y: float64(i * 20)}, eventManager)
		qt.Insert(enemy)
	}

	if !qt.Subdivided {
		t.Errorf("Quadtree should be subdivided")
	}

	qt.Clear()

	if qt.Subdivided {
		t.Errorf("Quadtree should not be subdivided after clear")
	}

	if len(qt.GetAllEntities()) != 0 {
		t.Errorf("Quadtree should be empty after clear")
	}
}

func TestCheckCollision(t *testing.T) {
	eventManager := &mocks.MockEventManager{}

	player := entity.NewPlayer(types.Vector2D{X: 10, Y: 10}, eventManager)
	enemy1 := entity.NewEnemy(types.Vector2D{X: 20, Y: 20}, eventManager)
	enemy2 := entity.NewEnemy(types.Vector2D{X: 50, Y: 50}, eventManager)

	if !checkCollision(player, enemy1) {
		t.Errorf("Expected collision between player and enemy1")
	}

	if checkCollision(player, enemy2) {
		t.Errorf("Did not expect collision between player and enemy2")
	}
}
