package entity

import (
	"testing"

	"github.com/ajkula/shmup/types"
	"github.com/stretchr/testify/assert"
)

func TestNewFormation(t *testing.T) {
	centerPos := types.Vector2D{X: 100, Y: 200}
	pattern := func(t float64) types.Vector2D { return types.Vector2D{X: t, Y: t} }
	formation := NewFormation("testFormation", pattern, centerPos)

	assert.Equal(t, "testFormation", formation.GetID())
	assert.Equal(t, centerPos, formation.GetCenterPos())
	assert.Empty(t, formation.GetEnemyIDs())
	assert.Equal(t, 0.0, formation.GetTime())
}

func TestFormationAddRemoveEnemy(t *testing.T) {
	formation := NewFormation("testFormation", nil, types.Vector2D{})

	formation.AddEnemy("enemy1")
	formation.AddEnemy("enemy2")
	assert.Len(t, formation.GetEnemyIDs(), 2)
	assert.Contains(t, formation.GetEnemyIDs(), "enemy1")
	assert.Contains(t, formation.GetEnemyIDs(), "enemy2")

	formation.RemoveEnemy("enemy1")
	assert.Len(t, formation.GetEnemyIDs(), 1)
	assert.NotContains(t, formation.GetEnemyIDs(), "enemy1")
	assert.Contains(t, formation.GetEnemyIDs(), "enemy2")
}

func TestFormationSetters(t *testing.T) {
	formation := NewFormation("testFormation", nil, types.Vector2D{X: 0, Y: 0})

	newPos := types.Vector2D{X: 150, Y: 250}
	formation.SetCenterPos(newPos)
	assert.Equal(t, newPos, formation.GetCenterPos())

	formation.SetTime(5.5)
	assert.Equal(t, 5.5, formation.GetTime())
}

func TestNewSpecificFormations(t *testing.T) {
	centerPos := types.Vector2D{X: 100, Y: 200}

	circleFormation := NewCircleFormation("circle", centerPos, 50)
	assert.NotNil(t, circleFormation)
	assert.Equal(t, "circle", circleFormation.GetID())
	assert.Equal(t, centerPos, circleFormation.GetCenterPos())

	sineWaveFormation := NewSineWaveFormation("sine", centerPos, 30, 0.1)
	assert.NotNil(t, sineWaveFormation)
	assert.Equal(t, "sine", sineWaveFormation.GetID())
	assert.Equal(t, centerPos, sineWaveFormation.GetCenterPos())

	vFormation := NewVFormation("v", centerPos, 40)
	assert.NotNil(t, vFormation)
	assert.Equal(t, "v", vFormation.GetID())
	assert.Equal(t, centerPos, vFormation.GetCenterPos())
}
