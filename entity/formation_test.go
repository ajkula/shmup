package entity

import (
	"testing"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/mocks"
	"github.com/ajkula/shmup/types"
)

type MockMovementPattern struct{}

// (wrong type for method Move)
// 		have Move([]types.Entity, float64)
// 		want Move([]types.FormationMember, float64)

func (m MockMovementPattern) Move(entities []types.FormationMember, elapsedTime float64) {
	for _, e := range entities {
		currentPos := e.GetPosition()
		e.SetPosition(types.Vector2D{X: currentPos.X, Y: currentPos.Y + 1})
	}
}

func TestNewFormation(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	position := types.Vector2D{X: 100, Y: 100}
	formationType := types.CircleFormation
	formation := NewFormation(formationType, nil, position, eventManager)

	if formation.GetPosition() != position {
		t.Errorf("NewFormation position: got %v, want %v", formation.GetPosition(), position)
	}
	if formation.GetFormationType() != formationType {
		t.Errorf("NewFormation type: got %v, want %v", formation.GetFormationType(), formation)
	}
	if len(formation.GetEntities()) != 0 {
		t.Errorf("NewFormation position: got %v, want 0", len(formation.GetEntities()))
	}
	eventManager.GetPublishedEvents()
	eventManager.Shutdown()
}

func TestFormationAddEntity(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	formation := NewFormation(types.LineFormation, nil, types.Vector2D{X: 0, Y: 0}, eventManager)
	enemy := NewEnemy(types.Vector2D{X: 10, Y: 10}, eventManager)

	eventManager.ClearPublishedEvents()
	formation.AddEntity(enemy)

	if len(formation.GetEntities()) != 1 {
		t.Errorf("Expected 1 entity in formation, got %d", len(formation.GetEntities()))
	}

	events := eventManager.GetPublishedEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event to be published, got %d", len(events))
	}
	if events[0].Type != interfaces.EnemyAddedToFormation {
		t.Errorf("Expected EnemyAddedToFormation event, got %v", events[0].Type)
	}
	if events[0].Data != enemy {
		t.Error("Expected the added enemy to be the event data")
	}
	eventManager.Shutdown()
}

func TestFormationRemoveEntity(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	mockPattern := MockMovementPattern{}
	formation := NewFormation(types.ColumnFormation, mockPattern, types.Vector2D{X: 0, Y: 0}, eventManager)
	enemy := NewEnemy(types.Vector2D{X: 10, Y: 10}, eventManager)

	eventManager.ClearPublishedEvents()
	formation.AddEntity(enemy)
	eventManager.ClearPublishedEvents() // clear events

	formation.RemoveEntity(enemy)

	if len(formation.GetEntities()) != 0 {
		t.Errorf("Expected 0 entities in formation after removal, got %d", len(formation.GetEntities()))
	}

	events := eventManager.GetPublishedEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event to be published, got %d", len(events))
	}
	if events[0].Type != interfaces.EnemyRemovedFromFormation {
		t.Errorf("Expected EnemyRemovedFromFormation event, got %v", events[0].Type)
	}
	if events[0].Data != enemy {
		t.Error("Expected the removed enemy to be the event data")
	}
	eventManager.Shutdown()
}

func TestFormationUpdate(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	mockPattern := MockMovementPattern{}
	formation := NewFormation(types.LineFormation, mockPattern, types.Vector2D{X: 0, Y: 0}, eventManager)
	enemy := NewEnemy(types.Vector2D{X: 10, Y: 10}, eventManager)
	formation.AddEntity(enemy)

	formation.Update(1.0)

	updatedPos := enemy.GetPosition()
	expectedPos := types.Vector2D{X: 10, Y: 11}
	if updatedPos != expectedPos {
		t.Errorf("Enemy position after update: got %v, want %v", updatedPos, expectedPos)
	}
	eventManager.Shutdown()
}

func TestFormationIsComplete(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	mockPattern := MockMovementPattern{}
	formation := NewFormation(types.LineFormation, mockPattern, types.Vector2D{X: 0, Y: 0}, eventManager)
	enemy := NewEnemy(types.Vector2D{X: 10, Y: 10}, eventManager)

	formation.AddEntity(enemy)
	formation.Update(0.16)
	if formation.IsComplete() {
		t.Errorf("Formation should not be complete when it has entities : %v", formation.enemies)
	}

	formation.RemoveEntity(enemy)
	formation.Update(0.16)
	if !formation.IsComplete() {
		t.Errorf("Formation should be complete when all entities are removed : %v", formation.enemies)
	}

	events := eventManager.GetPublishedEvents()
	if len(events) != 3 || events[2].Type != interfaces.FormationDestroyed {
		t.Errorf("Expected FormationCompleted event to be published")
	}
	eventManager.Shutdown()
}

func TestFormationSetPattern(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	initialPattern := MockMovementPattern{}
	formation := NewFormation(types.CircleFormation, initialPattern, types.Vector2D{X: 0, Y: 0}, eventManager)

	newPattern := MockMovementPattern{}
	formation.SetPattern(newPattern)

	if formation.GetPattern() != newPattern {
		t.Errorf("SetPattern did not update the pattern correctly")
	}
}
