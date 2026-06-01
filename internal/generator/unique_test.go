package generator

import (
	"sync"
	"testing"
)

func TestUniqueTrackerTryAdd(t *testing.T) {
	tracker := NewUniqueTracker()

	if !tracker.TryAdd("users.email", "alice@example.com") {
		t.Error("expected first add to succeed")
	}
	if tracker.TryAdd("users.email", "alice@example.com") {
		t.Error("expected duplicate add to fail")
	}
}

func TestUniqueTrackerAdd(t *testing.T) {
	tracker := NewUniqueTracker()

	if err := tracker.Add("users.id", int64(1)); err != nil {
		t.Fatal(err)
	}
	if err := tracker.Add("users.id", int64(1)); err == nil {
		t.Error("expected error for duplicate")
	}
}

func TestUniqueTrackerHas(t *testing.T) {
	tracker := NewUniqueTracker()
	tracker.Add("users.email", "bob@test.com")

	if !tracker.Has("users.email", "bob@test.com") {
		t.Error("expected Has to return true")
	}
	if tracker.Has("users.email", "nonexistent@test.com") {
		t.Error("expected Has to return false")
	}
	if tracker.Has("unknown.key", "anything") {
		t.Error("expected Has to return false for unknown key")
	}
}

func TestUniqueTrackerGenerate(t *testing.T) {
	tracker := NewUniqueTracker()

	counter := 0
	val, err := tracker.Generate("users.id", 100, func() any {
		counter++
		return counter
	})
	if err != nil {
		t.Fatal(err)
	}
	if val.(int) != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	val, err = tracker.Generate("users.id", 100, func() any {
		counter++
		return counter
	})
	if err != nil {
		t.Fatal(err)
	}
	if val.(int) != 2 {
		t.Errorf("expected 2, got %d", val)
	}
}

func TestUniqueTrackerGenerateExhausted(t *testing.T) {
	tracker := NewUniqueTracker()

	tracker.Add("test.key", "only_value")
	_, err := tracker.Generate("test.key", 5, func() any {
		return "only_value"
	})
	if err == nil {
		t.Error("expected error after exhausting retries")
	}
}

func TestUniqueTrackerReset(t *testing.T) {
	tracker := NewUniqueTracker()
	tracker.Add("users.email", "a@b.com")
	tracker.Reset("users.email")

	if tracker.Has("users.email", "a@b.com") {
		t.Error("expected value to be removed after reset")
	}
	if !tracker.TryAdd("users.email", "a@b.com") {
		t.Error("expected add to succeed after reset")
	}
}

func TestUniqueTrackerResetAll(t *testing.T) {
	tracker := NewUniqueTracker()
	tracker.Add("a.x", 1)
	tracker.Add("b.y", 2)
	tracker.ResetAll()

	if tracker.Has("a.x", 1) {
		t.Error("expected all values to be cleared")
	}
	if tracker.Has("b.y", 2) {
		t.Error("expected all values to be cleared")
	}
}

func TestUniqueTrackerConcurrent(t *testing.T) {
	tracker := NewUniqueTracker()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			tracker.Add("concurrent.key", n)
		}(i)
	}
	wg.Wait()

	for i := 0; i < 100; i++ {
		if !tracker.Has("concurrent.key", i) {
			t.Errorf("expected value %d to exist", i)
		}
	}
}

func TestUniqueTrackerSeparateKeys(t *testing.T) {
	tracker := NewUniqueTracker()
	tracker.Add("table_a.id", int64(1))
	tracker.Add("table_b.id", int64(1))

	if !tracker.Has("table_a.id", int64(1)) {
		t.Error("expected table_a.id to have 1")
	}
	if !tracker.Has("table_b.id", int64(1)) {
		t.Error("expected table_b.id to have 1")
	}
}
