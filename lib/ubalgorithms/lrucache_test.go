package ubalgorithms_test

import (
	"fmt"
	"testing"

	algorithms "github.com/kernelplex/ubase/lib/ubalgorithms"
)

func TestLRUCache_Remove_ExistingKey(t *testing.T) {
	lru := algorithms.NewLRUCache[int, string](10)
	lru.Put(1, "one")
	lru.Put(2, "two")

	// Remove existing key
	removed := lru.Remove(1)
	if !removed {
		t.Error("Expected Remove to return true for existing key")
	}

	// Verify removed
	_, found := lru.Get(1)
	if found {
		t.Error("Key should have been removed from cache")
	}

	// Verify other key still exists
	val, found := lru.Get(2)
	if !found || val != "two" {
		t.Error("Other key should still exist in cache")
	}
}

func TestLRUCache_Remove_NonExistingKey(t *testing.T) {
	lru := algorithms.NewLRUCache[int, string](10)
	lru.Put(1, "one")

	// Remove non-existing key
	removed := lru.Remove(2)
	if removed {
		t.Error("Expected Remove to return false for non-existing key")
	}

	// Verify original key still exists
	val, found := lru.Get(1)
	if !found || val != "one" {
		t.Error("Original key should still exist in cache")
	}
}

func TestPriorityCache_RetrievesValuesByKey(t *testing.T) {
	lru := algorithms.NewLRUCache[int, string](10)

	first := "one"
	second := "two"

	lru.Put(1, first)
	lru.Put(2, second)

	val, found := lru.Get(1)
	if !found {
		t.Errorf("Failed to retrieve item from queue.")
	}

	if val != first {
		t.Errorf("Incorrect value retrieved from cache wanted '%s' got '%s'", first, val)
	}

}

func TestPriorityCache_ReturnsFalseIfMissing(t *testing.T) {
	lru := algorithms.NewLRUCache[int, string](10)

	first := "one"

	lru.Put(1, first)

	val, found := lru.Get(2)
	if found {
		t.Errorf("Cache should not have returned a value. Got: %s", val)
	}

	if val != "" {
		t.Errorf("Incorrect value retrieved from cache wanted '%s' got '%s'", "", val)
	}

}

func TestPriorityCache_RollsOffOldestValue(t *testing.T) {
	lru := algorithms.NewLRUCache[int, string](10)

	for x := range 11 {
		lru.Put(x, fmt.Sprintf("Item %d", x))
	}

	// First item should have rolled off the cache.
	val, found := lru.Get(0)
	if found {
		t.Errorf("Cache should not have returned a value. Got: %s", val)
	}

	if val != "" {
		t.Errorf("Incorrect value retrieved from cache wanted '%s' got '%s'", "", val)
	}

	// Remaining items should still be in the cache.
	for x := 1; x < 11; x++ {
		expected := fmt.Sprintf("Item %d", x)
		val, found = lru.Get(x)

		if !found {
			t.Errorf("Failed to retrieve item from queue.")
		}
		if val != expected {
			t.Errorf("Incorrect value retrieved from cache wanted '%s' got '%s'", expected, val)
		}
	}
}
