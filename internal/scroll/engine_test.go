package scroll

import (
	"testing"
	"time"
)

func TestEngineSpeedBounds(t *testing.T) {
	engine := NewEngine(nil)
	t.Cleanup(engine.Stop)

	for i := 0; i < 100; i++ {
		engine.SpeedUp()
	}
	if got := engine.Speed(); got != DefaultMaxSpeed {
		t.Fatalf("expected max speed %v, got %v", DefaultMaxSpeed, got)
	}

	for i := 0; i < 100; i++ {
		engine.SpeedDown()
	}
	if got := engine.Speed(); got != DefaultMinSpeed {
		t.Fatalf("expected min speed %v, got %v", DefaultMinSpeed, got)
	}
}

func TestEngineToggle(t *testing.T) {
	engine := NewEngine(nil)
	t.Cleanup(engine.Stop)

	if engine.IsPlaying() {
		t.Fatal("expected engine to start paused")
	}

	engine.Play()
	if !engine.IsPlaying() {
		t.Fatal("expected engine to be playing after Play")
	}

	engine.Pause()
	if engine.IsPlaying() {
		t.Fatal("expected engine to be paused after Pause")
	}

	if !engine.Toggle() {
		t.Fatal("expected toggle to start playback")
	}
	if engine.Toggle() {
		t.Fatal("expected toggle to stop playback")
	}
}

func TestEngineEmitsDeltaWhenPlaying(t *testing.T) {
	deltaCh := make(chan float64, 1)

	engine := NewEngine(func(delta float64) {
		select {
		case deltaCh <- delta:
		default:
		}
	})
	t.Cleanup(engine.Stop)

	engine.Play()

	select {
	case delta := <-deltaCh:
		if delta <= 0 {
			t.Fatalf("expected a positive delta, got %v", delta)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("expected at least one delta emission while playing")
	}
}
