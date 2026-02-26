package scroll

import (
	"sync"
	"time"
)

const (
	DefaultMinSpeed = 20.0
	DefaultMaxSpeed = 300.0
	DefaultStep     = 20.0
	DefaultSpeed    = 20.0
	DefaultTickRate = 33 * time.Millisecond
)

type Engine struct {
	mu      sync.Mutex
	speed   float64
	min     float64
	max     float64
	step    float64
	playing bool

	ticker  *time.Ticker
	stopCh  chan struct{}
	onDelta func(float64)
}

func NewEngine(onDelta func(float64)) *Engine {
	engine := &Engine{
		speed:   DefaultSpeed,
		min:     DefaultMinSpeed,
		max:     DefaultMaxSpeed,
		step:    DefaultStep,
		ticker:  time.NewTicker(DefaultTickRate),
		stopCh:  make(chan struct{}),
		onDelta: onDelta,
	}

	go engine.loop()
	return engine
}

func (e *Engine) loop() {
	deltaSeconds := DefaultTickRate.Seconds()
	for {
		select {
		case <-e.stopCh:
			return
		case <-e.ticker.C:
			e.mu.Lock()
			playing := e.playing
			speed := e.speed
			e.mu.Unlock()

			if playing && e.onDelta != nil {
				e.onDelta(speed * deltaSeconds)
			}
		}
	}
}

func (e *Engine) Play() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.playing = true
}

func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.playing = false
}

func (e *Engine) Toggle() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.playing = !e.playing
	return e.playing
}

func (e *Engine) IsPlaying() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.playing
}

func (e *Engine) Speed() float64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.speed
}

func (e *Engine) SpeedUp() float64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.speed += e.step
	if e.speed > e.max {
		e.speed = e.max
	}
	return e.speed
}

func (e *Engine) SpeedDown() float64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.speed -= e.step
	if e.speed < e.min {
		e.speed = e.min
	}
	return e.speed
}

func (e *Engine) Stop() {
	e.ticker.Stop()
	close(e.stopCh)
}
