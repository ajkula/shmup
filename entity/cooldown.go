package entity

import (
	"sync"
	"time"
)

type ThreadSafeCooldown struct {
	mu          sync.RWMutex
	remaining   float64
	maxCooldown float64
}

func NewThreadSafeCooldown(cooldownSeconds float64) *ThreadSafeCooldown {
	return &ThreadSafeCooldown{
		remaining:   0.0,
		maxCooldown: cooldownSeconds,
	}
}

func (tsc *ThreadSafeCooldown) Update(deltaTime float64) {
	tsc.mu.Lock()
	defer tsc.mu.Unlock()

	if tsc.remaining > 0 {
		tsc.remaining -= deltaTime
		if tsc.remaining < 0 {
			tsc.remaining = 0
		}
	}
}

func (tsc *ThreadSafeCooldown) CanAct() bool {
	tsc.mu.RLock()
	defer tsc.mu.RUnlock()

	return tsc.remaining <= 0.001 // epsilon like original
}

func (tsc *ThreadSafeCooldown) TryAct() bool {
	tsc.mu.Lock()
	defer tsc.mu.Unlock()

	if tsc.remaining <= 0.001 {
		tsc.remaining = tsc.maxCooldown
		return true
	}
	return false
}

func (tsc *ThreadSafeCooldown) Reset() {
	tsc.mu.Lock()
	defer tsc.mu.Unlock()
	tsc.remaining = 0.0
}

func (tsc *ThreadSafeCooldown) GetRemainingCooldown() time.Duration {
	tsc.mu.RLock()
	defer tsc.mu.RUnlock()

	return time.Duration(tsc.remaining * float64(time.Second))
}

func (tsc *ThreadSafeCooldown) SetCooldownTime(seconds float64) {
	tsc.mu.Lock()
	defer tsc.mu.Unlock()
	tsc.maxCooldown = seconds
}
