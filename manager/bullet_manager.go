package manager

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/entity"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/registry"
	"github.com/ajkula/shmup/types"
)

type BulletManager struct {
	core.BaseSystem
	bullets       []types.GameEntity
	player        types.GameEntity
	eventManager  interfaces.EventManagerInterface
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
	isShutdown    int32
}

func NewBulletManager(eventManager interfaces.EventManagerInterface) *BulletManager {
	return &BulletManager{
		bullets:       make([]types.GameEntity, 0),
		eventManager:  eventManager,
		eventChannels: make(map[interfaces.EventType]<-chan interfaces.Event),
	}
}

func (bm *BulletManager) SetPlayer(player types.GameEntity) {
	bm.player = player
}

func (bm *BulletManager) Initialize(ctx context.Context) error {
	var err error
	err = bm.BaseSystem.Initialize(ctx)
	if err != nil {
		return err
	}

	eventTypes := []interfaces.EventType{
		interfaces.SystemTick,
		interfaces.BulletCreated,
		interfaces.BulletDestroyed,
		interfaces.PlayerShot,
		interfaces.EnemyShot,
		interfaces.BossShot,
		interfaces.BossDefeated,
		interfaces.GameReset,
	}

	for _, eventType := range eventTypes {
		bm.eventChannels[eventType], err = bm.eventManager.Subscribe(eventType)
		if err != nil {
			return fmt.Errorf("failed to initialize: %s", err)
		}
	}

	return nil
}

func (bm *BulletManager) Update(deltaTime float64) error {
	select {
	case <-bm.CTX.Done():
		return bm.CTX.Err()
	default:
		bm.processAllAvailableEvents()
		bm.updateBullets()
		return nil
	}
}

func (bm *BulletManager) processAllAvailableEvents() {
	if atomic.LoadInt32(&bm.isShutdown) == 1 {
		return
	}

	for eventType, ch := range bm.eventChannels {
		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					return
				}
				bm.handleEvent(eventType, evt)
			default:
				goto nextChannel
			}
		}
	nextChannel:
	}
}

func (bm *BulletManager) handleEvent(eventType interfaces.EventType, evt interfaces.Event) {
	switch eventType {
	case interfaces.BulletCreated, interfaces.BulletDestroyed:
		bm.handleBulletEvent(evt)
	case interfaces.PlayerShot:
		bm.handlePlayerShot(evt)
	case interfaces.EnemyShot, interfaces.BossShot:
		bm.handleEnemyShot(evt)
	case interfaces.BossDefeated:
		bm.destroyAllEnemyBullets()
	case interfaces.GameReset:
		bm.handleGameReset()
	}
}

func (bm *BulletManager) updateBullets() {
	aliveBullets := make([]types.GameEntity, 0, len(bm.bullets))
	for _, bullet := range bm.bullets {
		if err := bullet.Update(core.FixedDeltaTime); err != nil {
			continue
		}
		if bullet.IsAlive() {
			aliveBullets = append(aliveBullets, bullet)
		}
	}
	bm.bullets = aliveBullets
}

func (bm *BulletManager) handlePlayerShot(evt interfaces.Event) {
	shooter, ok := evt.Data.(types.GameEntity)
	if !ok {
		return
	}

	pos := shooter.GetPosition()
	bullet := entity.NewBullet(pos.X+16, pos.Y, false, bm.eventManager)
	bm.bullets = append(bm.bullets, bullet)
	bm.eventManager.Publish(interfaces.BulletCreated, bullet)
}

func (bm *BulletManager) handleEnemyShot(evt interfaces.Event) {
	if patternEvent, ok := evt.Data.(interfaces.PatternShootEvent); ok {
		bm.handlePatternShot(patternEvent)
		return
	}

	shooter, ok := evt.Data.(types.GameEntity)
	if !ok {
		return
	}

	pos := shooter.GetPosition()

	if bm.player != nil && bm.player.IsAlive() {
		playerPos := bm.player.GetPosition()
		dx := playerPos.X - pos.X
		dy := playerPos.Y - pos.Y

		length := math.Sqrt(dx*dx + dy*dy)
		if length > 0 {
			direction := types.Vector2D{X: dx / length, Y: dy / length}
			bullet := entity.NewBulletWithDirection(
				pos.X+16, pos.Y+32,
				direction,
				200,  // speed
				true, // isEnemy = true
				bm.eventManager,
			)
			bm.bullets = append(bm.bullets, bullet)
			bm.eventManager.Publish(interfaces.BulletCreated, bullet)
			return
		}
	}

	bullet := entity.NewBullet(pos.X+16, pos.Y+32, true, bm.eventManager)
	bm.bullets = append(bm.bullets, bullet)
	bm.eventManager.Publish(interfaces.BulletCreated, bullet)
}

func (bm *BulletManager) handlePatternShot(patternEvent interfaces.PatternShootEvent) {
	shooterPos := patternEvent.Shooter.GetPosition()

	var targetPos *types.Vector2D
	if bm.player != nil && bm.player.IsAlive() {
		playerPos := bm.player.GetPosition()
		targetPos = &playerPos
	}

	shots := patternEvent.Pattern.GenerateShots(shooterPos, targetPos)

	for _, shot := range shots {
		bullet := entity.NewBulletWithDirection(
			shot.Position.X, shot.Position.Y,
			shot.Direction, shot.Speed,
			true, // isEnemy = true
			bm.eventManager,
		)
		bm.bullets = append(bm.bullets, bullet)
		bm.eventManager.Publish(interfaces.BulletCreated, bullet)
	}
}

func (bm *BulletManager) handleBulletEvent(evt interfaces.Event) {
	if bullet, ok := evt.Data.(types.GameEntity); ok {
		switch evt.Type {
		case interfaces.BulletCreated:
			for _, existing := range bm.bullets {
				if existing == bullet {
					return
				}
			}
			bm.bullets = append(bm.bullets, bullet)
		case interfaces.BulletDestroyed:
			for i, b := range bm.bullets {
				if b == bullet {
					bm.bullets = append(bm.bullets[:i], bm.bullets[i+1:]...)
					break
				}
			}
		}
	}
}

func (bm *BulletManager) GetRenderableEntities() []types.Renderable {
	renderables := make([]types.Renderable, 0, len(bm.bullets))
	for _, bullet := range bm.bullets {
		if bullet.IsAlive() {
			renderables = append(renderables, bullet)
		}
	}
	return renderables
}

func (bm *BulletManager) GetBulletCount() int {
	return len(bm.bullets)
}

func (bm *BulletManager) GetBullets() []types.GameEntity {
	bullets := make([]types.GameEntity, len(bm.bullets))
	copy(bullets, bm.bullets)
	return bullets
}

func (bm *BulletManager) destroyAllEnemyBullets() {
	// Destroy all enemy bullets when boss is defeated
	// This prevents the player from dying after the boss is defeated
	playerBullets := make([]types.GameEntity, 0)
	for _, bullet := range bm.bullets {
		if b, ok := bullet.(*entity.Bullet); ok {
			if !b.IsEnemyBullet() {
				// Keep player bullets
				playerBullets = append(playerBullets, bullet)
			}
			// Enemy bullets are discarded
		}
	}
	bm.bullets = playerBullets
	fmt.Println("[BulletManager] Destroyed all enemy bullets after boss defeat")
}

func (bm *BulletManager) handleGameReset() {
	bm.ClearAllBullets()
}

func (bm *BulletManager) ClearAllBullets() {
	// Clear all bullets when game is reset
	bm.bullets = bm.bullets[:0]
	fmt.Println("[BulletManager] All bullets cleared - game reset")
}

func (bm *BulletManager) Shutdown() {
	if !atomic.CompareAndSwapInt32(&bm.isShutdown, 0, 1) {
		return
	}

	for eventType, ch := range bm.eventChannels {
		bm.eventManager.Unsubscribe(eventType, ch)
	}
	bm.eventChannels = nil
	bm.bullets = nil
}

var _ registry.EntityProvider = (*BulletManager)(nil)
var _ core.System = (*BulletManager)(nil)
