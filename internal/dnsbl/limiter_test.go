package dnsbl

import (
	"testing"
	"time"
)

func TestSlidingWindow(t *testing.T) {
	t.Parallel()
	w := &slidingWindow{window: time.Hour, max: 3}
	now := time.Unix(1000000, 0)
	if !w.allow(now) || !w.allow(now.Add(time.Minute)) || !w.allow(now.Add(2*time.Minute)) {
		t.Fatal("expected 3 allows")
	}
	if w.allow(now.Add(3 * time.Minute)) {
		t.Fatal("expected 4th deny within window")
	}
}

func TestMinuteCounter(t *testing.T) {
	t.Parallel()
	m := minuteCounter{perMinute: 2}
	now := time.Unix(3600, 0)
	if !m.allow(now) || !m.allow(now) {
		t.Fatal("expected 2 allows same minute")
	}
	if m.allow(now) {
		t.Fatal("expected 3rd deny")
	}
	nextMin := now.Add(time.Minute)
	if !m.allow(nextMin) {
		t.Fatal("expected reset next minute")
	}
}

func TestMinuteCounterUnlimited(t *testing.T) {
	t.Parallel()
	m := minuteCounter{perMinute: 0}
	for i := 0; i < 50; i++ {
		if !m.allow(time.Now()) {
			t.Fatal("unlimited should always allow")
		}
	}
}
