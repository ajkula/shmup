package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type BulletManager struct {
	core.BaseSystem
	bullets       []types.GameEntity
	eventManager  interfaces.EventManagerInterface
	mu            sync.RWMutex
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
}

func NewBulletManager(eventManager interfaces.EventManagerInterface) *BulletManager {
	return &BulletManager{
		bullets:      make([]types.GameEntity, 0),
		eventManager: eventManager,
	}
}

func (bm *BulletManager) Initialize(ctx context.Context) error {
	var err error
	err = bm.BaseSystem.Initialize(ctx)
	if err != nil {
		return err
	}

	eventTypes := []interfaces.EventType{
		interfaces.BulletCreated,
		interfaces.BulletDestroyed,
	}

	bm.eventChannels = make(map[interfaces.EventType]<-chan interfaces.Event)
	for _, eventType := range eventTypes {
		bm.eventChannels[eventType], err = bm.eventManager.Subscribe(eventType)
		if err != nil {
			return fmt.Errorf("failedto initilize: %s", err)
		}
	}

	return nil
}

func (bm *BulletManager) Update(deltaTime float64) error {
	select {
	case <-bm.CTX.Done():
		return bm.CTX.Err()
	default:
		eventsToProcess := bm.gatherEvents()

		bm.mu.Lock()
		defer bm.mu.Unlock()

		for _, evt := range eventsToProcess {
			bm.handleEvent(evt)
		}

		for _, bullet := range bm.bullets {
			if err := bullet.Update(deltaTime); err != nil {
				return err
			}
		}

		return nil
	}
}

func (bm *BulletManager) gatherEvents() []interfaces.Event {
	var events []interfaces.Event
	for _, ch := range bm.eventChannels {
		// Limit the number of processed events per chan to avoid infinite loop
		for i := 0; i < 100; i++ {
			select {
			case evt, ok := <-ch:
				if !ok {
					break
				}
				events = append(events, evt)
			default:
				// nothing
			}
		}
	}
	return events
}

func (bm *BulletManager) handleEvent(evt interfaces.Event) {
	if bullet, ok := evt.Data.(types.GameEntity); ok {
		switch evt.Type {
		case interfaces.BulletCreated:
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

func (bm *BulletManager) Draw(screen *ebiten.Image) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	for _, bullet := range bm.bullets {
		bullet.Draw(screen)
	}
}

func (bm *BulletManager) AddBullet(bullet types.GameEntity) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.bullets = append(bm.bullets, bullet)
}

func (bm *BulletManager) RemoveBullet(bullet types.GameEntity) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	for i, b := range bm.bullets {
		if b == bullet {
			bm.bullets = append(bm.bullets[:i], bm.bullets[i+1:]...)
			break
		}
	}
}

func (bm *BulletManager) Shutdown() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	for eventType, ch := range bm.eventChannels {
		bm.eventManager.Unsubscribe(eventType, ch)
	}
	bm.eventChannels = nil
	bm.bullets = nil
}

var _ core.System = (*BulletManager)(nil)
